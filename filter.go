package main

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"google.golang.org/api/gmail/v1"
)

// filterfile defines a set of filter objects.
type filterfile struct {
	Filter []filter
}

// filter defines a filter object.
type filter struct {
	Query   string
	QueryOr []string
	Archive bool
	Read    bool
	Delete  bool
	Label   string
}

func (f filter) toGmailFilters() ([]gmail.Filter, error) {
	// Convert the filter into a gmail filters.
	if len(f.Query) > 0 && len(f.QueryOr) > 0 {
		return nil, errors.New("cannot have both a query and a queryOr")
	}

	filters := []gmail.Filter{
		{
			Action:   &gmail.FilterAction{},
			Criteria: &gmail.FilterCriteria{},
		},
	}
	return filters, nil
}

func (f filter) addFilter() error {
	// Convert the filter into a gmail filter.
	filters, err := f.toGmailFilters()
	if err != nil {
		return err
	}

	// Add the filters.
	for _, filter := range filters {
		if _, err := api.Users.Settings.Filters.Create(gmailUser).Do(gmailUser, &filter); err != nil {
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
