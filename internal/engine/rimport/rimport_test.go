package rimport

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mbrt/gmailctl/internal/engine/apply"
	"github.com/mbrt/gmailctl/internal/engine/filter"
)

// Regression test for https://github.com/mbrt/gmailctl/issues/456:
// a downloaded filter using raw exclusion syntax with parentheses (e.g.
// "-(some-term)") must round-trip unchanged through import and export.
// Previously, needsEscape() and filter.quote() checked different character
// sets, so a value with parentheses (but no spaces or quotes) was imported
// as "not escaped" and then re-quoted on export, silently inverting the
// intended exclusion.
func TestImportPreservesRawParens(t *testing.T) {
	fs := filter.Filters{
		{
			Criteria: filter.Criteria{
				Subject: "-(some-term)",
			},
			Action: filter.Actions{
				Archive: true,
			},
		},
	}

	cfg, err := Import(fs, nil)
	require.Nil(t, err)
	require.Len(t, cfg.Rules, 1)

	node := cfg.Rules[0].Filter
	assert.Equal(t, "-(some-term)", node.Subject)
	assert.True(t, node.IsEscaped, "value with parens must be marked as already escaped")

	// The re-exported filter must match the original criteria exactly.
	pres, err := apply.FromConfig(cfg)
	require.Nil(t, err)
	require.Len(t, pres.Filters, 1)
	assert.Equal(t, fs[0].Criteria, pres.Filters[0].Criteria)
}

func TestNeedsEscape(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"plain", "term", false},
		{"space", "some term", true},
		{"single quote", "some'term", true},
		{"double quote", `some"term`, true},
		{"parens", "-(some-term)", true},
		{"braces", "{some-term}", true},
		{"tab", "some\tterm", true},
		{"plus without at", "foo+bar", true},
		{"plus with at", "foo+bar@example.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, needsEscape(tt.in))
		})
	}
}
