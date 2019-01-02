package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/gmail/v1"
)

// filterfile defines a set of filter objects.
type filterfile struct {
	Filter []filter
}

// filter defines a filter object.
type filter struct {
	Query             string
	QueryOr           []string
	Archive           bool
	Read              bool
	Delete            bool
	ToMe              bool
	ArchiveUnlessToMe bool
	Label             string
	ForwardTo         string
}

func (f filter) toGmailFilters(labels *labelMap) ([]gmail.Filter, error) {
	// Convert the filter into a gmail filters.
	if len(f.Query) > 0 && len(f.QueryOr) > 0 {
		return nil, errors.New("cannot have both a query and a queryOr")
	}

	if len(f.QueryOr) > 0 {
		// Create the OR query.
		f.Query = strings.Join(f.QueryOr, " OR ")
	}

	if len(f.Query) < 1 {
		return nil, errors.New("query or queryOr cannot be empty")
	}

	action := gmail.FilterAction{
		AddLabelIds:    []string{},
		RemoveLabelIds: []string{},
	}
	if len(f.Label) > 0 {
		// Create the label if it does not exist.
		labelID, err := labels.createLabelIfDoesNotExist(f.Label)
		if err != nil {
			return nil, err
		}
		action.AddLabelIds = append(action.AddLabelIds, labelID)
	}

	action.RemoveLabelIds = []string{}
	if f.Archive && !f.ArchiveUnlessToMe {
		action.RemoveLabelIds = append(action.RemoveLabelIds, "INBOX")
	}

	if f.Read {
		action.RemoveLabelIds = append(action.RemoveLabelIds, "UNREAD")
	}

	if f.Delete {
		action.AddLabelIds = append(action.AddLabelIds, "TRASH")
	}

	if len(f.ForwardTo) > 0 {
		action.Forward = f.ForwardTo
	}

	criteria := gmail.FilterCriteria{
		Query: f.Query,
	}
	if f.ToMe || f.ArchiveUnlessToMe {
		criteria.To = "me"
	}

	filter := gmail.Filter{
		Action:   &action,
		Criteria: &criteria,
	}
	filters := []gmail.Filter{
		filter,
	}

	// If we need to archive unless to them, then add the additional filter.
	if f.ArchiveUnlessToMe {
		// Copy the filter.
		archiveIfNotToMeFilter := filter
		archiveIfNotToMeFilter.Criteria = &gmail.FilterCriteria{
			Query:        f.Query,
			To:           "",
			NegatedQuery: "to:me",
		}

		// Copy the action.
		archiveAction := action
		// Archive it.
		archiveAction.RemoveLabelIds = append(action.RemoveLabelIds, "INBOX")
		archiveIfNotToMeFilter.Action = &archiveAction

		// Append the extra filter.
		filters = append(filters, archiveIfNotToMeFilter)
	}

	return filters, nil
}

func (f filter) addFilter(labels *labelMap) error {
	// Convert the filter into a gmail filter.
	filters, err := f.toGmailFilters(labels)
	if err != nil {
		return err
	}

	// Add the filters.
	for _, fltr := range filters {
		logrus.WithFields(logrus.Fields{
			"action":   fmt.Sprintf("%#v", fltr.Action),
			"criteria": fmt.Sprintf("%#v", fltr.Criteria),
		}).Debug("adding Gmail filter")
		if _, err := api.Users.Settings.Filters.Create(gmailUser, &fltr).Do(); err != nil {
			return fmt.Errorf("creating filter failed: %v", err)
		}
	}

	return nil
}

func decodeFile(file string) ([]filter, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("reading filter file %s failed: %v", file, err)
	}

	var ff filterfile
	if _, err := toml.Decode(string(b), &ff); err != nil {
		return nil, fmt.Errorf("decoding toml failed: %v", err)
	}

	return ff.Filter, nil
}

func downloadExistingFilters(l *labelMap) error {
	labels := *l
	filterList, err := api.Users.Settings.Filters.List(gmailUser).Do()
	if err != nil {
		return fmt.Errorf("listing filters failed: %v", err)
	}

	backupFilterJsonFile, err := ioutil.TempFile("./", "backupFilterJson")
	if err != nil {
		return fmt.Errorf("failed to create backup filter json file: %v", err)
	}
	defer backupFilterJsonFile.Close()

	backupFilterTomlFile, err := ioutil.TempFile("./", "backupFilterToml")
	if err != nil {
		return fmt.Errorf("failed to create backup filter json file: %v", err)
	}
	defer backupFilterTomlFile.Close()

	filterJson, err := json.Marshal(&filterList)
	if err != nil {
		return fmt.Errorf("Failed to marshal filterList")
	}

	if _, err := backupFilterJsonFile.Write(filterJson); err != nil {
		return fmt.Errorf("failed to write filterList to backup filter file: %v", err)
	}

	for _, googleFilter := range filterList.Filter {
		Label := ""
		ForwardTo := ""
		// queryOr := []string{}
		Archive := false
		Read := false
		ToMe := false
		Delete := false
		ArchiveUnlessToMe := false
		query := ""

		// Start toml string
		filterString := "[[filter]]" + "\n"
		if len(googleFilter.Criteria.From) > 0 {
			query += "from:(" + googleFilter.Criteria.From + ") "
		}

		if len(googleFilter.Criteria.Subject) > 0 {
			query += "subject:(" + googleFilter.Criteria.Subject + ") "
		}

		if len(googleFilter.Criteria.To) > 0 {
			query += "to:" + googleFilter.Criteria.To + " "
		}

		query += googleFilter.Criteria.Query

		filterString += "query = \"\"\" " + query + " \"\"\"" + "\n"

		for _, label := range googleFilter.Action.RemoveLabelIds {
			if label == "INBOX" {
				Archive = true
				filterString += "archive = " + strconv.FormatBool(Archive) + "\n"
				if googleFilter.Criteria.NegatedQuery == "to:me" {
					ArchiveUnlessToMe = true
					filterString += "archiveUnlessToMe = " + strconv.FormatBool(ArchiveUnlessToMe) + "\n"
				}
			}
			if label == "UNREAD" {
				Read = true
				filterString += "read = " + strconv.FormatBool(Read) + "\n"
			}
		}

		for _, label := range googleFilter.Action.AddLabelIds {
			if label == "TRASH" {
				Delete = true
				filterString += "delete = " + strconv.FormatBool(Delete) + "\n"
			}
			for labelName, labelId := range labels {
				if label == labelId {
					Label = labelName
					filterString += "label = \"" + Label + "\"\n"
				}
			}
		}

		if googleFilter.Criteria.To == "me" {
			ToMe = true
			filterString += "ToMe = " + strconv.FormatBool(ToMe) + "\n"
		}

		if len(googleFilter.Action.Forward) > 0 {
			ForwardTo = "\"" + googleFilter.Action.Forward + "\""
			filterString += "ForwardTo = " + ForwardTo + "\n"
		}

		filterString += "\n"

		if _, err := backupFilterTomlFile.WriteString(filterString); err != nil {
			return fmt.Errorf("failed to write localFilter to backup toml file: %v", err)
		}
	}
	return nil
}

func deleteExistingFilters() error {
	// Get current filters for the user.
	l, err := api.Users.Settings.Filters.List(gmailUser).Do()
	if err != nil {
		return fmt.Errorf("listing filters failed: %v", err)
	}

	// Iterate over the filters.
	for _, f := range l.Filter {
		// Delete the filter.
		if err := api.Users.Settings.Filters.Delete(gmailUser, f.Id).Do(); err != nil {
			return fmt.Errorf("deleting filter id %s failed: %v", f.Id, err)
		}
	}

	return nil
}
