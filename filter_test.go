package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/gmail/v1"
)

func TestFilterToGmailFilters(t *testing.T) {
	testCases := map[string]struct {
		orig     filter
		expected []gmail.Filter
	}{
		"archive unless to me": {
			orig: filter{
				Query:             "list:coreos-dev@googlegroups.com",
				Label:             "Mailing Lists/coreos-dev",
				ArchiveUnlessToMe: true,
			},

			expected: []gmail.Filter{
				{
					Action: &gmail.FilterAction{
						AddLabelIds: []string{"1"},
					},
					Criteria: &gmail.FilterCriteria{
						Query: "list:coreos-dev@googlegroups.com",
						To:    "me",
					},
				},
				{
					Action: &gmail.FilterAction{
						AddLabelIds:    []string{"1"},
						RemoveLabelIds: []string{"INBOX"},
					},
					Criteria: &gmail.FilterCriteria{
						NegatedQuery: "to:me",
						Query:        "list:coreos-dev@googlegroups.com",
					},
				},
			},
		},
	}

	labels := &labelMap{
		strings.ToLower("Mailing Lists/coreos-dev"): "1",
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			filters, err := tc.orig.toGmailFilters(labels)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expected, filters); len(diff) > 1 {
				t.Fatalf("got diff: %s", diff)
			}
		})
	}
}
