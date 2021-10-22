package apply_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/mbrt/gmailctl/internal/fakegmail"
	"github.com/mbrt/gmailctl/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestOne(t *testing.T) {
	svc := fakegmail.NewService(context.Background(), t)
	api := api.NewFromService(svc)
	// err := api.AddLabels(label.Labels{
	// 	{
	// 		Name:  "Label1",
	// 		Color: &label.Color{Background: "red", Text: "blue"},
	// 	},
	// 	{Name: "Label2"},
	// })
	// assert.Nil(t, err)

	fs, err := api.ListLabels()
	assert.Nil(t, err)
	fmt.Print(fs)
}
