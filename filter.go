package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
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
			return fmt.Errorf("creating filter [%#v] failed: %v", fltr, err)
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

func exportExistingFilters(file string) error {
	fmt.Print("exporting existing filters...\n")

	filters, err := getExistingFilters()
	if err != nil {
		return fmt.Errorf("error downloading existing filters: %v", err)
	}

	var ff filterfile
	for _, f := range filters {
		// We could get duplicate filters, so it's best to remove them.
		existingFilter := findExistingFilter(ff.Filter, f.Query)

		// Since we can't return nil on a struct or compary it to something empty,
		// check if the query exists. If not then consider it not found.
		if existingFilter.Query != "" {
			// Duplicate filters can only exist if the ArchiveUnlessToMe is set.
			// So we can simply reset everything and just set the ArchiveUnlessToMe flag to true.
			existingFilter.Archive = false
			existingFilter.Delete = false
			existingFilter.ToMe = false
			existingFilter.ArchiveUnlessToMe = true
		} else {
			ff.Filter = append(ff.Filter, f)
		}
	}

	return writeFiltersToFile(ff, file)
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

func getExistingFilters() ([]filter, error) {
	gmailFilters, err := api.Users.Settings.Filters.List(gmailUser).Do()
	if err != nil {
		return nil, err
	}

	labels, err := getLabelMapOnID()
	if err != nil {
		return nil, err
	}

	var filters []filter

	for _, gmailFilter := range gmailFilters.Filter {
		var f filter

		if gmailFilter.Criteria.Query > "" {
			f.Query = gmailFilter.Criteria.Query

			if gmailFilter.Criteria.To == "me" {
				f.ToMe = true
			}

			if len(gmailFilter.Action.AddLabelIds) > 0 {
				labelID := gmailFilter.Action.AddLabelIds[0]
				if labelID == "TRASH" {
					f.Delete = true
				} else {
					labelName, ok := labels[labelID]
					if ok {
						f.Label = labelName
					}
				}
			}

			if len(gmailFilter.Action.RemoveLabelIds) > 0 {
				for _, labelID := range gmailFilter.Action.RemoveLabelIds {
					if labelID == "UNREAD" {
						f.Read = true
					} else if labelID == "INBOX" {
						if gmailFilter.Criteria.NegatedQuery == "to:me" {
							f.ArchiveUnlessToMe = true
						} else {
							f.Archive = true
						}
					}
				}
			}
		}

		filters = append(filters, f)
	}

	return filters, nil
}

func writeFiltersToFile(ff filterfile, file string) error {
	exportFile, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("error exporting filters: %v", err)
	}

	writer := bufio.NewWriter(exportFile)
	encoder := toml.NewEncoder(writer)
	encoder.Indent = ""

	if err := encoder.Encode(ff); err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	fmt.Printf("Exported %d filters\n", len(ff.Filter))

	return nil
}

func findExistingFilter(filters []filter, query string) filter {
	for _, f := range filters {
		if f.Query == query {
			return f
		}
	}

	return filter{}
}
