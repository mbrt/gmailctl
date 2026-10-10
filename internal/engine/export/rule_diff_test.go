package export_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mbrt/gmailctl/internal/engine/config/v1alpha3"
	"github.com/mbrt/gmailctl/internal/engine/export"
	"github.com/mbrt/gmailctl/internal/engine/export/xml"
	"github.com/mbrt/gmailctl/internal/engine/filter"
)

func TestRuleDiffExporter(t *testing.T) {
	fixedTime := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	author := v1alpha3.Author{
		Name:  "Test User",
		Email: "user@example.com",
	}

	exp := export.NewRuleDiffExporterWithXMLExporter(
		xml.NewWithTime(func() time.Time { return fixedTime }),
	)

	f1 := filter.Filter{
		Criteria: filter.Criteria{From: "existing@example.com"},
		Action:   filter.Actions{MarkRead: true},
	}
	f1Upstream := filter.Filter{
		ID:       "id-existing",
		Criteria: filter.Criteria{From: "existing@example.com"},
		Action:   filter.Actions{MarkRead: true},
	}
	fNew := filter.Filter{
		Criteria: filter.Criteria{From: "new@example.com"},
		Action:   filter.Actions{Star: true},
	}
	fModified := filter.Filter{
		Criteria: filter.Criteria{From: "modified@example.com"},
		Action:   filter.Actions{Archive: true},
	}
	fModifiedOld := filter.Filter{
		ID:       "id-modified",
		Criteria: filter.Criteria{From: "modified@example.com"},
		Action:   filter.Actions{Archive: false, Star: true},
	}

	t.Run("exports only new and modified filters", func(t *testing.T) {
		local := filter.Filters{f1, fNew, fModified}
		upstream := filter.Filters{f1Upstream, fModifiedOld}

		var buf bytes.Buffer
		err := exp.ExportNew(author, local, upstream, &buf)
		require.NoError(t, err)

		out := buf.String()
		assert.Contains(t, out, "<title>Mail Filters</title>")
		assert.Contains(t, out, "<name>Test User</name>")
		assert.Contains(t, out, "<email>user@example.com</email>")

		// Must contain new and modified
		assert.Contains(t, out, `name="from" value="new@example.com"`)
		assert.Contains(t, out, `name="shouldStar" value="true"`)
		assert.Contains(t, out, `name="from" value="modified@example.com"`)
		assert.Contains(t, out, `name="shouldArchive" value="true"`)

		// Must not contain existing unchanged filter
		assert.NotContains(t, out, `name="from" value="existing@example.com"`)
	})

	t.Run("exports empty feed when no differences", func(t *testing.T) {
		local := filter.Filters{f1}
		upstream := filter.Filters{f1Upstream}

		var buf bytes.Buffer
		err := exp.ExportNew(author, local, upstream, &buf)
		require.NoError(t, err)

		out := buf.String()
		assert.Contains(t, out, "<title>Mail Filters</title>")
		assert.NotContains(t, out, "<entry>")
	})

	t.Run("default exporter constructor", func(t *testing.T) {
		defaultExp := export.NewRuleDiffExporter()
		local := filter.Filters{fNew}

		var buf bytes.Buffer
		err := defaultExp.ExportNew(author, local, nil, &buf)
		require.NoError(t, err)

		assert.Contains(t, buf.String(), `name="from" value="new@example.com"`)
	})
}
