package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	gmailv1 "google.golang.org/api/gmail/v1"

	"github.com/mbrt/gmailctl/pkg/filter"
	"github.com/mbrt/gmailctl/pkg/gmail"
)

func TestImportActions(t *testing.T) {
	filters := []*gmailv1.Filter{
		{
			Action: &gmailv1.FilterAction{
				AddLabelIds: []string{
					labelIDTrash,
					labelIDImportant,
					labelIDStar,
					labelIDCategoryUpdates,
				},
				RemoveLabelIds: []string{
					labelIDInbox,
					labelIDUnread,
					labelIDSpam,
				},
			},
			Criteria: &gmailv1.FilterCriteria{
				From: "foo@bar.com",
			},
		},
	}
	imported, err := DefaulImporter().Import(filters, emptyLabelMap())
	expected := filter.Filters{
		{
			Action: filter.Actions{
				Archive:       true,
				Delete:        true,
				MarkRead:      true,
				Star:          true,
				MarkNotSpam:   true,
				MarkImportant: true,
				Category:      gmail.CategoryUpdates,
			},
			Criteria: filter.Criteria{
				From: "foo@bar.com",
			},
		},
	}

	assert.Nil(t, err)
	assert.Equal(t, imported, expected)
}

func TestImportCriteria(t *testing.T) {
	filters := []*gmailv1.Filter{
		{
			Action: &gmailv1.FilterAction{
				AddLabelIds:    []string{labelIDTrash},
				RemoveLabelIds: []string{},
			},
			Criteria: &gmailv1.FilterCriteria{
				From:    "foo@bar.com",
				To:      "baz@zuz.it",
				Subject: "baz",
				Query:   "my query",
			},
		},
	}
	imported, err := DefaulImporter().Import(filters, emptyLabelMap())
	expected := filter.Filters{
		{
			Action: filter.Actions{
				Delete: true,
			},
			Criteria: filter.Criteria{
				From:    "foo@bar.com",
				To:      "baz@zuz.it",
				Subject: "baz",
				Query:   "my query",
			},
		},
	}

	assert.Nil(t, err)
	assert.Equal(t, imported, expected)
}

func TestImportLabels(t *testing.T) {
	filters := []*gmailv1.Filter{
		{
			Action: &gmailv1.FilterAction{
				AddLabelIds: []string{
					labelIDCategoryForums,
					"label1",
				},
				RemoveLabelIds: []string{},
			},
			Criteria: &gmailv1.FilterCriteria{
				From: "foo@bar.com",
			},
		},
	}
	lmap := NewLabelMap(map[string]string{
		"label1": "MyLabel",
		"label2": "NewLabel",
	})

	imported, err := DefaulImporter().Import(filters, lmap)
	expected := filter.Filters{
		{
			Action: filter.Actions{
				Category: gmail.CategoryForums,
				AddLabel: "MyLabel",
			},
			Criteria: filter.Criteria{
				From: "foo@bar.com",
			},
		},
	}

	assert.Nil(t, err)
	assert.Equal(t, imported, expected)

	// Test not existing label
	filters = []*gmailv1.Filter{
		{
			Action: &gmailv1.FilterAction{
				AddLabelIds: []string{
					labelIDCategoryForums,
					"labelXXX",
				},
				RemoveLabelIds: []string{},
			},
			Criteria: &gmailv1.FilterCriteria{
				From: "foo@bar.com",
			},
		},
	}
	_, err = DefaulImporter().Import(filters, lmap)
	assert.NotNil(t, err)
}

func TestImportBad(t *testing.T) {
	// Importing filters with missing pieces doesn't cause crashes.
	filters := []*gmailv1.Filter{
		{
			Action: &gmailv1.FilterAction{
				AddLabelIds: []string{labelIDTrash},
			},
			Criteria: nil,
		},
		{
			Action: &gmailv1.FilterAction{
				AddLabelIds: []string{labelIDTrash},
			},
			Criteria: &gmailv1.FilterCriteria{
				From: "foo@bar.com",
			},
		},
	}
	imported, err := DefaulImporter().Import(filters, emptyLabelMap())
	assert.NotNil(t, err)
	assert.Len(t, imported, 1)

	filters = []*gmailv1.Filter{
		{
			Action: nil,
			Criteria: &gmailv1.FilterCriteria{
				From: "foo@bar.com",
			},
		},
		{
			Action: &gmailv1.FilterAction{
				AddLabelIds: []string{labelIDTrash},
			},
			Criteria: &gmailv1.FilterCriteria{
				From: "foo@bar.com",
			},
		},
	}
	imported, err = DefaulImporter().Import(filters, emptyLabelMap())
	assert.NotNil(t, err)
	assert.Len(t, imported, 1)
}
