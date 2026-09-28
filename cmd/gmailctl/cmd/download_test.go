package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	papply "github.com/mbrt/gmailctl/internal/engine/apply"
	"github.com/mbrt/gmailctl/internal/engine/filter"
	"github.com/mbrt/gmailctl/internal/engine/label"
	"github.com/mbrt/gmailctl/internal/engine/rimport"
)

func TestDownloadLabels(t *testing.T) {
	upstream := papply.GmailConfig{
		Labels: label.Labels{{Name: "work"}},
		Filters: filter.Filters{{
			Criteria: filter.Criteria{From: "foo@example.com"},
			Action:   filter.Actions{AddLabel: "work"},
		}},
	}

	tests := []struct {
		name          string
		includeLabels bool
	}{
		{"with labels", true},
		{"without labels", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := importConfig(upstream, tt.includeLabels)
			require.Nil(t, err)
			var buf bytes.Buffer
			err = rimport.MarshalJsonnet(cfg, &buf, downloadHeader)
			require.Nil(t, err)

			out := buf.String()
			assert.Equal(t, tt.includeLabels, len(cfg.Labels) > 0)
			// Top-level labels section (the rule's actions.labels is nested deeper).
			assert.Equal(t, tt.includeLabels, strings.Contains(out, "\n  labels: [\n"), out)
			assert.Equal(t, tt.includeLabels, strings.Contains(out, "labels management is optional"), out)
			// The rule referencing the label is kept regardless.
			assert.Contains(t, out, "        labels: [\n          \"work\"\n")
		})
	}
}
