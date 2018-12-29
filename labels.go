package main

import (
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"google.golang.org/api/gmail/v1"
)

type labelMap map[string]string

func getLabelMap() (labelMap, error) {
	// Get the labels for the user and map its name to its ID.
	l, err := api.Users.Labels.List(gmailUser).Do()
	if err != nil {
		return nil, fmt.Errorf("listing labels failed: %v", err)
	}

	labels := labelMap{}
	for _, label := range l.Labels {
		labels[strings.ToLower(label.Name)] = label.Id
	}

	return labels, nil
}

func (m *labelMap) createLabelIfDoesNotExist(name string) (string, error) {
	// De reference the pointer so we can index.
	labels := *m

	// Try to find the label.
	id, ok := labels[strings.ToLower(name)]
	if ok {
		// We found the label.
		return id, nil
	}

	// Create the label if it does not exist.
	label, err := api.Users.Labels.Create(gmailUser, &gmail.Label{Name: name}).Do()
	if err != nil {
		return "", fmt.Errorf("creating label %s failed: %v", name, err)
	}
	logrus.Infof("Created label: %s", name)

	// Update our label map.
	labels[strings.ToLower(name)] = label.Id
	m = &labels
	return label.Id, nil
}
