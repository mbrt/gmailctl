package label

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mbrt/gmailctl/internal/engine/filter"
)

func TestInvalid(t *testing.T) {
	cases := []struct {
		name   string
		labels Labels
	}{
		{
			"unnamed",
			Labels{{Name: ""}},
		},
		{
			"starts with slash",
			Labels{{Name: "/foobar"}},
		},
		{
			"ends with slash",
			Labels{{Name: "foobar/"}},
		},
		{
			"duplicates",
			Labels{
				{Name: "abc"},
				{Name: "abc"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.labels.Validate()
			assert.NotNil(t, err)
		})
	}
}

func TestValid(t *testing.T) {
	cases := []struct {
		name   string
		labels Labels
	}{
		{"empty", nil},
		{
			"single",
			Labels{{Name: "foobar"}},
		},
		{
			"sub-labels",
			Labels{
				{Name: "abc/def"},
				{Name: "abc"},
				{Name: "abc/def/ghi"},
				{Name: "another"},
			},
		},
		{
			"missing prefix",
			Labels{
				{Name: "abc/def"},
				{Name: "ab"},
			},
		},
		{
			"missing prefix 2",
			Labels{
				{Name: "abc"},
				{Name: "abc/def/ghi"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.labels.Validate()
			assert.Nil(t, err)
		})
	}
}

func TestValidateUsed(t *testing.T) {
	d := LabelsDiff{
		Removed: Labels{{Name: "foo"}},
	}
	fs := filter.Filters{
		{
			Criteria: filter.Criteria{To: "foobar"},
			Action:   filter.Actions{AddLabel: "foo"},
		},
	}
	err := Validate(d, fs)
	assert.NotNil(t, err)
}
