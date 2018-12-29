package main

import (
	"fmt"
	"strings"
)

func getLabelMap() (map[string]string, error) {
	// Get the labels for the user and map its name to its ID.
	l, err := api.Users.Labels.List(gmailUser).Do()
	if err != nil {
		return nil, fmt.Errorf("listing labels failed: %v", err)
	}

	labels := map[string]string{}
	for _, label := range l.Labels {
		labels[strings.ToLower(label.Name)] = label.Id
	}

	return labels, nil
}
