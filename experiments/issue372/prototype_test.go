package main

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mbrt/gmailctl/internal/engine/api"
	"github.com/mbrt/gmailctl/internal/engine/apply"
	cfg "github.com/mbrt/gmailctl/internal/engine/config/v1alpha3"
	"github.com/mbrt/gmailctl/internal/engine/filter"
	"github.com/mbrt/gmailctl/internal/engine/parser"
	"github.com/mbrt/gmailctl/internal/engine/rimport"
	"github.com/mbrt/gmailctl/internal/fakegmail"
)

func mustGenerate(t *testing.T, config cfg.Config) candidates {
	t.Helper()
	c, err := generate(config)
	require.Nil(t, err)
	return c
}

func TestCompactIssueExampleAndActions(t *testing.T) {
	config, err := readConfig("examples/compact.jsonnet")
	require.Nil(t, err)
	require.True(t, config.Settings.Compact)
	c := mustGenerate(t, config.Config)
	require.Len(t, c.Normal, 4)
	require.Equal(t, filter.Filters{
		{Criteria: filter.Criteria{Query: "{cc:list1@example.com list:list1.example.com}"}, Action: filter.Actions{Archive: true, AddLabel: "work"}},
		{Criteria: filter.Criteria{Query: "{cc:list1@example.com list:list1.example.com}"}, Action: filter.Actions{AddLabel: "lists"}},
	}, c.Compact)
}

func TestCompactKeepsNestedGrouping(t *testing.T) {
	c := mustGenerate(t, cfg.Config{Version: cfg.Version, Rules: []cfg.Rule{{
		Filter: cfg.FilterNode{Or: []cfg.FilterNode{
			{And: []cfg.FilterNode{{From: "a"}, {Not: &cfg.FilterNode{Subject: "b"}}}},
			{To: "c"},
		}},
		Actions: cfg.Actions{Archive: true},
	}}})
	require.Len(t, c.Normal, 2)
	require.Len(t, c.Compact, 1)
	require.Equal(t, "{to:c (from:a -subject:b)}", c.Compact[0].Criteria.Query)
}

func TestCompactGuards(t *testing.T) {
	longSubject := func(length int) cfg.FilterNode {
		return cfg.FilterNode{Or: []cfg.FilterNode{{From: "a"}, {Subject: strings.Repeat("x", length)}}}
	}
	// The rendering is {from:a subject:<value>}, which adds 17 bytes.
	for _, tc := range []struct {
		name string
		node cfg.FilterNode
		want int
	}{
		{"raw query", cfg.FilterNode{Or: []cfg.FilterNode{{From: "a"}, {Query: "subject:b to:c"}}}, 2},
		{"escaped", cfg.FilterNode{Or: []cfg.FilterNode{{From: "a"}, {Subject: "-(b)", IsEscaped: true}}}, 2},
		{"at byte budget", longSubject(983), 1},
		{"over byte budget", longSubject(984), 2},
		{"same field already grouped", cfg.FilterNode{Or: []cfg.FilterNode{{From: "a"}, {From: "b"}}}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mustGenerate(t, cfg.Config{Version: cfg.Version, Rules: []cfg.Rule{{Filter: tc.node, Actions: cfg.Actions{Archive: true}}}})
			require.Len(t, c.Compact, tc.want)
		})
	}
	for _, argumentCount := range []int{17, 18} {
		t.Run(fmt.Sprintf("node budget with %d from arguments", argumentCount), func(t *testing.T) {
			node := cfg.FilterNode{Or: []cfg.FilterNode{{Subject: "topic"}}}
			for i := 0; i < argumentCount; i++ {
				node.Or = append(node.Or, cfg.FilterNode{From: fmt.Sprintf("a%d", i)})
			}
			c := mustGenerate(t, cfg.Config{Version: cfg.Version, Rules: []cfg.Rule{{Filter: node, Actions: cfg.Actions{Archive: true}}}})
			if argumentCount == 17 { // 17 args + subject + OR node = 19.
				require.Len(t, c.Compact, 1)
			} else { // 20 meets the existing split threshold.
				require.Equal(t, c.Normal, c.Compact)
			}
		})
	}
}

func TestCountsDeduplicateExpandedActions(t *testing.T) {
	config, err := readConfig("examples/normal.jsonnet")
	require.Nil(t, err)
	rule := config.Rules[0]
	for i := 0; i < 1000; i++ {
		config.Rules = append(config.Rules, rule)
	}
	c := mustGenerate(t, config.Config)
	require.Len(t, c.Normal, 4)
	require.Len(t, c.Compact, 2)
	require.False(t, chooseHysteresis(c, nil, false).Compact)
}

func TestCompactionDoesNotLoseSharedFilterSavings(t *testing.T) {
	c := mustGenerate(t, overlappingRules())
	require.Len(t, c.Normal, 5)
	require.Equal(t, 6, c.ProposedCompact)
	require.Equal(t, c.Normal, c.Compact)
	require.Zero(t, c.CompactedRules)
}

func TestHysteresisBoundariesAndRepeat(t *testing.T) {
	for _, n := range []int{749, 750, 900, 901} {
		for _, previouslyCompact := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d previous compact=%t", n, previouslyCompact), func(t *testing.T) {
				c := mustGenerate(t, synthetic(n, "stable"))
				up := chooseToggle(previouslyCompact).Filters(c)
				d := chooseHysteresis(c, up, false)
				want := previouslyCompact
				if n < 750 {
					want = false
				} else if n > 900 {
					want = true
				}
				require.Equal(t, want, d.Compact)
				installed := d.Filters(c)
				repeat := chooseHysteresis(c, installed, false)
				require.Equal(t, d.Compact, repeat.Compact)
				require.Zero(t, changeCount(installed, repeat.Filters(c)).Total())
				require.Equal(t, n > 900, chooseHysteresis(c, installed, true).Compact)
			})
		}
	}
}

func TestInferenceIsNotHistoricalState(t *testing.T) {
	c := mustGenerate(t, synthetic(850, "old"))
	// No existing mode exists, but minimizing additions selects compact.
	require.True(t, chooseHysteresis(c, nil, false).Compact)
	rewritten := mustGenerate(t, synthetic(850, "new"))
	// A complete rewrite loses evidence of the old normal mode.
	require.True(t, chooseHysteresis(rewritten, c.Normal, false).Compact)
	for _, compactPairs := range []int{212, 213} {
		up := append(append(filter.Filters{}, c.Compact[:compactPairs]...), c.Normal[compactPairs*2:]...)
		require.Equal(t, compactPairs == 213, chooseHysteresis(c, up, false).Compact)
	}
	require.False(t, chooseHysteresis(candidates{}, nil, false).Compact)
}

func TestSequentialChurn(t *testing.T) {
	rows, err := thresholdTrace()
	require.Nil(t, err)
	require.Equal(t, 1275, rows[0].OnPeak)
	require.Equal(t, changes{451, 898}, rows[2].AutoDelta)
	require.Equal(t, changes{0, 2}, rows[3].AutoDelta)
	require.Equal(t, changes{748, 375}, rows[5].AutoDelta)
	require.Equal(t, changes{0, 1}, rows[5].OnDelta)
	require.Equal(t, changes{898, 449}, rows[3].ExportDelta)
	for _, row := range rows {
		require.False(t, row.Off.Compact)
		require.True(t, row.On.Compact)
	}
}

func TestNoticesExplainRegroupingAndStayQuietOnNoDiff(t *testing.T) {
	before := mustGenerate(t, synthetic(898, "stable"))
	after := mustGenerate(t, synthetic(902, "stable"))
	var out bytes.Buffer
	d := chooseHysteresis(after, before.Normal, false)
	planNotice(&out, after, d, before.Normal, false)
	require.Contains(t, out.String(), "normal count 902 exceeds 900")
	require.Contains(t, out.String(), "replaces 898 installed filters")
	require.Contains(t, out.String(), "peak at 1349")
	out.Reset()
	planNotice(&out, after, d, after.Compact, false)
	require.Empty(t, out.String())
	out.Reset()
	planNotice(&out, after, d, nil, false)
	require.NotContains(t, out.String(), "replaces")
	out.Reset()
	planNotice(&out, after, chooseToggle(false), before.Normal, false)
	require.Contains(t, out.String(), "Setting settings.compact to true would reduce it to 451")
	require.NotContains(t, out.String(), "replaces")
}

func TestNormalAndChangeCountsMatchProduction(t *testing.T) {
	paths, err := filepath.Glob("../../testdata/valid/*.jsonnet")
	require.Nil(t, err)
	require.NotEmpty(t, paths)
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			config, err := readConfig(path)
			require.Nil(t, err)
			c := mustGenerate(t, config.Config)
			rules, err := parser.Parse(config.Config)
			require.Nil(t, err)
			existing, err := filter.FromRules(rules)
			require.Nil(t, err)
			require.Equal(t, filterSet(existing), filterSet(c.Normal))
			for i := range existing {
				existing[i].ID = fmt.Sprintf("server-id-%d", i)
			}
			diff, err := filter.Diff(existing, c.Compact, false, 3, false)
			require.Nil(t, err)
			require.Equal(t, changes{len(diff.Added), len(diff.Removed)}, changeCount(existing, c.Compact))
		})
	}
}

func TestFakeGmailApplyDownloadAndRepeat(t *testing.T) {
	config := synthetic(850, "stable")
	c := mustGenerate(t, config)
	gapi := api.NewFromService(fakegmail.NewService(context.Background(), t))
	// Start with compact filters installed, as after a previous threshold crossing.
	require.Nil(t, gapi.AddFilters(c.Compact))
	upstream, err := gapi.ListFilters()
	require.Nil(t, err)
	d := chooseHysteresis(c, upstream, false)
	require.True(t, d.Compact)
	require.Zero(t, changeCount(upstream, d.Filters(c)).Total())

	// One ordinary edit stays compact in the band and uses the actual apply layer.
	config.Rules[0].Filter.Or[1].Subject = "new topic"
	c = mustGenerate(t, config)
	d = chooseHysteresis(c, upstream, false)
	require.True(t, d.Compact)
	fdiff, err := filter.Diff(upstream, d.Filters(c), false, 3, false)
	require.Nil(t, err)
	require.Len(t, fdiff.Added, 1)
	require.Len(t, fdiff.Removed, 1)
	require.Nil(t, apply.Apply(apply.ConfigDiff{FiltersDiff: fdiff}, gapi, false))
	upstream, err = gapi.ListFilters()
	require.Nil(t, err)
	repeat := chooseHysteresis(c, upstream, false)
	require.Zero(t, changeCount(upstream, repeat.Filters(c)).Total())

	downloaded, err := rimport.Import(upstream, nil)
	require.Nil(t, err)
	imported := mustGenerate(t, downloaded)
	require.Zero(t, changeCount(upstream, imported.Normal).Total())
	require.Zero(t, changeCount(upstream, imported.Compact).Total())
}
