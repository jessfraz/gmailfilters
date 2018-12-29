package main

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
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
