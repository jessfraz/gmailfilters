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
						AddLabelIds:    []string{"1"},
						RemoveLabelIds: []string{},
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
		"archive unless to me with OR": {
			orig: filter{
				QueryOr:           []string{"list:xdg-app@lists.freedesktop.org", "list:flatpak@lists.freedesktop.org"},
				Label:             "Mailing Lists/xdg-apps",
				ArchiveUnlessToMe: true,
			},
			expected: []gmail.Filter{
				{
					Action: &gmail.FilterAction{
						AddLabelIds:    []string{"2"},
						RemoveLabelIds: []string{},
					},
					Criteria: &gmail.FilterCriteria{
						Query: "list:xdg-app@lists.freedesktop.org OR list:flatpak@lists.freedesktop.org",
						To:    "me",
					},
				},
				{
					Action: &gmail.FilterAction{
						AddLabelIds:    []string{"2"},
						RemoveLabelIds: []string{"INBOX"},
					},
					Criteria: &gmail.FilterCriteria{
						NegatedQuery: "to:me",
						Query:        "list:xdg-app@lists.freedesktop.org OR list:flatpak@lists.freedesktop.org",
					},
				},
			},
		},
		"delete": {
			orig: filter{
				QueryOr: []string{"to:plans@tripit.com", "to:receipts@expensify.com"},
				Delete:  true,
			},
			expected: []gmail.Filter{
				{
					Action: &gmail.FilterAction{
						AddLabelIds:    []string{"TRASH"},
						RemoveLabelIds: []string{},
					},
					Criteria: &gmail.FilterCriteria{
						Query: "to:plans@tripit.com OR to:receipts@expensify.com",
					},
				},
			},
		},
		"archive and read": {
			orig: filter{
				QueryOr: []string{"to:plans@tripit.com", "to:receipts@expensify.com"},
				Archive: true,
				Read:    true,
			},
			expected: []gmail.Filter{
				{
					Action: &gmail.FilterAction{
						AddLabelIds:    []string{},
						RemoveLabelIds: []string{"INBOX", "UNREAD"},
					},
					Criteria: &gmail.FilterCriteria{
						Query: "to:plans@tripit.com OR to:receipts@expensify.com",
					},
				},
			},
		},
	}

	labels := &labelMap{
		strings.ToLower("Mailing Lists/coreos-dev"): "1",
		strings.ToLower("Mailing Lists/xdg-apps"):   "2",
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
