package fakegmail_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mbrt/gmailctl/internal/fakegmail"
	"github.com/mbrt/gmailctl/pkg/api"
	"github.com/mbrt/gmailctl/pkg/label"
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
