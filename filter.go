package main

import (
	"errors"
	"fmt"
	"io/ioutil"
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
	if f.Archive {
		action.RemoveLabelIds = append(action.RemoveLabelIds, "INBOX")
	}

	if f.Read {
		action.RemoveLabelIds = append(action.RemoveLabelIds, "UNREAD")
	}

	if f.Delete {
		action.AddLabelIds = append(action.AddLabelIds, "TRASH")
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
		filter.Criteria = &gmail.FilterCriteria{
			Query:        f.Query,
			To:           "",
			NegatedQuery: "to:me",
		}

		// Archive it.
		action.RemoveLabelIds = append(action.RemoveLabelIds, "INBOX")
		filter.Action = &action

		// Append the extra filter.
		filters = append(filters, filter)
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
