package fakegmail_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mbrt/gmailctl/internal/engine/api"
	"github.com/mbrt/gmailctl/internal/engine/filter"
	"github.com/mbrt/gmailctl/internal/engine/gmail"
	"github.com/mbrt/gmailctl/internal/engine/label"
	"github.com/mbrt/gmailctl/internal/fakegmail"
)

func TestLabels(t *testing.T) {
	svc := fakegmail.NewService(context.Background(), t)
	api := api.NewFromService(svc)

	// Add.
	err := api.AddLabels(label.Labels{
		{
			Name:  "Label1",
			Color: &label.Color{Background: "red", Text: "blue"},
		},
		{Name: "Label2"},
	})
	assert.Nil(t, err)

	// List.
	ls, err := api.ListLabels()
	assert.Nil(t, err)
	assert.Len(t, ls, 2)

	// Add duplicate.
	err = api.AddLabels(label.Labels{{Name: "Label2"}})
	assert.NotNil(t, err)

	// Delete.
	err = api.DeleteLabels([]string{ls[0].ID})
	assert.Nil(t, err)
	ls, err = api.ListLabels()
	assert.Nil(t, err)
	assert.Len(t, ls, 1)

	// Update.
	ls[0].Color = &label.Color{
		Background: "green",
		Text:       "blue",
	}
	err = api.UpdateLabels(ls)
	assert.Nil(t, err)
	ls, err = api.ListLabels()
	assert.Nil(t, err)
	assert.Len(t, ls, 1)
	assert.Equal(t, "green", ls[0].Color.Background)
	assert.Equal(t, "blue", ls[0].Color.Text)
}

func TestFilters(t *testing.T) {
	svc := fakegmail.NewService(context.Background(), t)
	api := api.NewFromService(svc)

	// Add label.
	err := api.AddLabels(label.Labels{{Name: "label1"}})
	assert.Nil(t, err)

	// Add.
	err = api.AddFilters(filter.Filters{
		{
			Criteria: filter.Criteria{
				From: "address@mail.com",
			},
			Action: filter.Actions{
				Category:      gmail.CategoryPersonal,
				MarkImportant: true,
			},
		},
		{
			Criteria: filter.Criteria{
				Subject: "foo",
			},
			Action: filter.Actions{
				AddLabel: "label1",
			},
		},
	})
	assert.Nil(t, err)

	// List.
	fs, err := api.ListFilters()
	assert.Nil(t, err)
	assert.Len(t, fs, 2)

	// Add duplicate.
	err = api.AddFilters(filter.Filters{
		{
			Criteria: filter.Criteria{
				Subject: "foo",
			},
			Action: filter.Actions{
				AddLabel: "label1",
			},
		},
	})
	assert.NotNil(t, err)

	// Add with non existing label.
	err = api.AddFilters(filter.Filters{
		{
			Criteria: filter.Criteria{
				Subject: "bar",
			},
			Action: filter.Actions{
				AddLabel: "this-does-not-exist",
			},
		},
	})
	assert.NotNil(t, err)

	// Delete.
	err = api.DeleteFilters([]string{fs[0].ID})
	assert.Nil(t, err)
	fs, err = api.ListFilters()
	assert.Nil(t, err)
	assert.Len(t, fs, 1)
}
