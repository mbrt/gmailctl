package xml_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mbrt/gmailctl/internal/engine/config/v1alpha3"
	"github.com/mbrt/gmailctl/internal/engine/export/xml"
	"github.com/mbrt/gmailctl/internal/engine/filter"
	"github.com/mbrt/gmailctl/internal/engine/gmail"
)

func TestXMLExport(t *testing.T) {
	fixedTime := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	author := v1alpha3.Author{
		Name:  "Test Author",
		Email: "author@example.com",
	}
	exp := xml.NewWithTime(func() time.Time { return fixedTime })

	t.Run("empty filters", func(t *testing.T) {
		var buf bytes.Buffer
		err := exp.Export(author, nil, &buf)
		require.NoError(t, err)

		out := buf.String()
		assert.True(t, strings.HasPrefix(out, "<?xml version="))
		assert.Contains(t, out, "<feed ")
		assert.Contains(t, out, "<title>Mail Filters</title>")
		assert.Contains(t, out, "<name>Test Author</name>")
		assert.Contains(t, out, "<email>author@example.com</email>")
		assert.NotContains(t, out, "<entry>")
	})

	t.Run("all criteria and action properties", func(t *testing.T) {
		filters := filter.Filters{
			{
				Criteria: filter.Criteria{
					From:    "sender@example.com",
					To:      "recipient@example.com",
					Subject: "Hello World",
					Query:   "has:attachment -label:spam",
				},
				Action: filter.Actions{
					Archive:          true,
					Delete:           true,
					MarkImportant:    true,
					MarkNotImportant: false,
					MarkRead:         true,
					MarkNotSpam:      true,
					Star:             true,
					AddLabel:         "custom-label",
					Forward:          "fwd@example.com",
					Category:         gmail.CategoryPromotions,
				},
			},
		}

		var buf bytes.Buffer
		err := exp.Export(author, filters, &buf)
		require.NoError(t, err)

		out := buf.String()
		assert.Contains(t, out, `name="from" value="sender@example.com"`)
		assert.Contains(t, out, `name="to" value="recipient@example.com"`)
		assert.Contains(t, out, `name="subject" value="Hello World"`)
		assert.Contains(t, out, `name="hasTheWord" value="has:attachment -label:spam"`)

		assert.Contains(t, out, `name="shouldArchive" value="true"`)
		assert.Contains(t, out, `name="shouldTrash" value="true"`)
		assert.Contains(t, out, `name="shouldAlwaysMarkAsImportant" value="true"`)
		assert.Contains(t, out, `name="shouldMarkAsRead" value="true"`)
		assert.Contains(t, out, `name="shouldNeverSpam" value="true"`)
		assert.Contains(t, out, `name="shouldStar" value="true"`)
		assert.Contains(t, out, `name="label" value="custom-label"`)
		assert.Contains(t, out, `name="forwardTo" value="fwd@example.com"`)
		assert.Contains(t, out, `name="smartLabelToApply" value="^smartlabel_promo"`)
	})

	t.Run("mark not important", func(t *testing.T) {
		filters := filter.Filters{
			{
				Criteria: filter.Criteria{From: "automated@example.com"},
				Action:   filter.Actions{MarkNotImportant: true},
			},
		}

		var buf bytes.Buffer
		err := exp.Export(author, filters, &buf)
		require.NoError(t, err)

		out := buf.String()
		assert.Contains(t, out, `name="shouldNeverMarkAsImportant" value="true"`)
		assert.NotContains(t, out, `name="shouldAlwaysMarkAsImportant"`)
	})

	t.Run("invalid category returns error", func(t *testing.T) {
		filters := filter.Filters{
			{
				Criteria: filter.Criteria{From: "bad@example.com"},
				Action:   filter.Actions{Category: gmail.Category("unknown_category")},
			},
		}

		var buf bytes.Buffer
		err := exp.Export(author, filters, &buf)
		require.Error(t, err)
	})
}
