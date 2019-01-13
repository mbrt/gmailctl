package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	gmailv1 "google.golang.org/api/gmail/v1"

	"github.com/mbrt/gmailctl/pkg/filter"
	"github.com/mbrt/gmailctl/pkg/gmail"
)

func emptyLabelMap() DefaultLabelMap {
	return NewDefaultLabelMap(map[string]string{})
}

func TestExportActions(t *testing.T) {
	filters := filter.Filters{
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
	exported, err := DefaulExporter().Export(filters, emptyLabelMap())
	expected := []*gmailv1.Filter{
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

	assert.Nil(t, err)
	assert.Equal(t, expected, exported)
}

func TestExportCriteria(t *testing.T) {
	filters := filter.Filters{
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
	exported, err := DefaulExporter().Export(filters, emptyLabelMap())
	expected := []*gmailv1.Filter{
		{
			Action: &gmailv1.FilterAction{
				AddLabelIds: []string{labelIDTrash},
			},
			Criteria: &gmailv1.FilterCriteria{
				From:    "foo@bar.com",
				To:      "baz@zuz.it",
				Subject: "baz",
				Query:   "my query",
			},
		},
	}

	assert.Nil(t, err)
	assert.Equal(t, expected, exported)
}

func TestExportNoActions(t *testing.T) {
	filters := filter.Filters{
		{
			Criteria: filter.Criteria{
				From: "foo@bar.com",
			},
		},
	}
	_, err := DefaulExporter().Export(filters, emptyLabelMap())
	assert.NotNil(t, err)
}

func TestExportNoCriteria(t *testing.T) {
	filters := filter.Filters{
		{
			Action: filter.Actions{
				Category: gmail.CategoryForums,
			},
		},
	}
	_, err := DefaulExporter().Export(filters, emptyLabelMap())
	assert.NotNil(t, err)
}

func TestExportLabels(t *testing.T) {
	filters := filter.Filters{
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
	lmap := NewDefaultLabelMap(map[string]string{
		"label1": "MyLabel",
		"label2": "NewLabel",
	})

	exported, err := DefaulExporter().Export(filters, lmap)
	expected := []*gmailv1.Filter{
		{
			Action: &gmailv1.FilterAction{
				AddLabelIds: []string{
					labelIDCategoryForums,
					"label1",
				},
			},
			Criteria: &gmailv1.FilterCriteria{
				From: "foo@bar.com",
			},
		},
	}

	assert.Nil(t, err)
	assert.Equal(t, expected, exported)

	// Test not existing label
	filters = filter.Filters{
		{
			Action: filter.Actions{
				AddLabel: "NonExisting",
			},
			Criteria: filter.Criteria{
				From: "foo@bar.com",
			},
		},
	}
	_, err = DefaulExporter().Export(filters, emptyLabelMap())
	assert.NotNil(t, err)
}
