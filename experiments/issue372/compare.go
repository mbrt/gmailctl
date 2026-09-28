package main

import (
	"fmt"
	"io"
	"path/filepath"

	cfg "github.com/mbrt/gmailctl/internal/engine/config/v1alpha3"
	"github.com/mbrt/gmailctl/internal/engine/filter"
)

// Every pair has different fields and distinct values, so normal generation
// produces two filters and compact generation produces one. An odd count
// includes one ordinary single-field rule. Values remain stable across sizes.
func synthetic(normalCount int, namespace string) cfg.Config {
	c := cfg.Config{Version: cfg.Version}
	for i := 0; i < normalCount/2; i++ {
		c.Rules = append(c.Rules, cfg.Rule{
			Filter: cfg.FilterNode{Or: []cfg.FilterNode{
				{From: fmt.Sprintf("%s-sender-%04d@example.com", namespace, i)},
				{Subject: fmt.Sprintf("%s-topic-%04d", namespace, i)},
			}},
			Actions: cfg.Actions{Archive: true},
		})
	}
	if normalCount%2 == 1 {
		c.Rules = append(c.Rules, cfg.Rule{
			Filter:  cfg.FilterNode{To: namespace + "-single@example.com"},
			Actions: cfg.Actions{Archive: true},
		})
	}
	return c
}

func overlappingRules() cfg.Config {
	c := cfg.Config{Version: cfg.Version}
	for _, from := range []string{"a", "b"} {
		for _, subject := range []string{"x", "y", "z"} {
			c.Rules = append(c.Rules, cfg.Rule{
				Filter:  cfg.FilterNode{Or: []cfg.FilterNode{{From: from}, {Subject: subject}}},
				Actions: cfg.Actions{Archive: true},
			})
		}
	}
	return c
}

type traceRow struct {
	Name                         string
	Normal, Compact              int
	Auto, Off, On                decision
	AutoDelta, OffDelta          changes
	OnDelta                      changes
	AutoPeak, OnPeak, AutoExport int
	ExportDelta                  changes
}

func thresholdTrace() ([]traceRow, error) {
	initial, err := generate(synthetic(850, "stable"))
	if err != nil {
		return nil, err
	}
	autoUp, offUp, onUp := initial.Normal, initial.Normal, initial.Normal
	steps := []struct {
		name  string
		count int
	}{
		{"Upgrade / enable toggle", 850},
		{"Grow within band", 898},
		{"Cross upper threshold", 902},
		{"Drop back into band", 898},
		{"Reach lower boundary", 750},
		{"Cross lower threshold", 748},
		{"Grow back into band", 752},
		{"Cross upper threshold again", 902},
	}
	var rows []traceRow
	for _, step := range steps {
		c, err := generate(synthetic(step.count, "stable"))
		if err != nil {
			return nil, err
		}
		auto := chooseHysteresis(c, autoUp, false)
		off, on := chooseToggle(false), chooseToggle(true)
		autoFilters := auto.Filters(c)
		exported := chooseHysteresis(c, nil, true).Filters(c)
		row := traceRow{
			Name: step.name, Normal: len(c.Normal), Compact: len(c.Compact),
			Auto: auto, Off: off, On: on,
			AutoDelta:  changeCount(autoUp, autoFilters),
			OffDelta:   changeCount(offUp, off.Filters(c)),
			OnDelta:    changeCount(onUp, on.Filters(c)),
			AutoExport: len(exported), ExportDelta: changeCount(autoFilters, exported),
		}
		row.AutoPeak = len(autoUp) + row.AutoDelta.Added
		row.OnPeak = len(onUp) + row.OnDelta.Added
		rows = append(rows, row)
		// Assume each desired state has been installed, even when the current
		// create-before-delete apply order would need separate capacity handling.
		autoUp, offUp, onUp = autoFilters, off.Filters(c), on.Filters(c)
	}
	return rows, nil
}

func compare(w io.Writer, fixtureDir string) error {
	fmt.Fprintln(w, "# Issue #372: measured prototype comparison")
	fmt.Fprintln(w, "\nThese counts use the real Jsonnet/parser/filter pipeline plus a shared experimental compactor. Synthetic cases are controlled examples, not a sample of user configurations. No Gmail requests are made.")
	fmt.Fprintln(w, "\nPolicies: hysteresis enters compact mode above 900 normal filters, exits below 750, and chooses fewer additions + removals inside the band (normal on ties). Export uses only the >900 threshold. The toggle obeys settings.compact, default false, for both online planning and export.")
	fmt.Fprintln(w, "\n## Sequential edits")
	fmt.Fprintln(w, "\nAll strategies start with the same 850 normal filters installed. Toggle=true opts in on the first row. Subsequent rows use each strategy's previous output as installed state. Each cell reports selected count and additions/removals. The trace assumes successful installation between steps; peaks exceeding 1000 are explicitly infeasible under today's apply order.")
	fmt.Fprintln(w, "\n| Edit | Normal candidate | Hysteresis | Toggle=false | Toggle=true | Hysteresis apply peak | Toggle=true apply peak |")
	fmt.Fprintln(w, "|---|---:|---|---|---|---:|---:|")
	rows, err := thresholdTrace()
	if err != nil {
		return err
	}
	for _, r := range rows {
		selected := r.Normal
		if r.Auto.Compact {
			selected = r.Compact
		}
		fmt.Fprintf(w, "| %s | %d | %s: %d; %s | %d; %s | %d; %s | %d | %d |\n", r.Name, r.Normal, r.Auto.Mode(), selected, r.AutoDelta, r.Normal, r.OffDelta, r.Compact, r.OnDelta, r.AutoPeak, r.OnPeak)
	}
	fmt.Fprintln(w, "\n## Offline export after the same edits")
	fmt.Fprintln(w, "\nDifferences compare export with the successfully installed result of online planning. Both toggle settings always export their selected representation.")
	fmt.Fprintln(w, "\n| Edit | Hysteresis export count | Difference from installed filters |")
	fmt.Fprintln(w, "|---|---:|---|")
	for _, r := range rows {
		fmt.Fprintf(w, "| %s | %d | %s |\n", r.Name, r.AutoExport, r.ExportDelta)
	}

	fmt.Fprintln(w, "\n## Mode inference under ambiguous history")
	base, err := generate(synthetic(850, "old"))
	if err != nil {
		return err
	}
	rewritten, err := generate(synthetic(850, "new"))
	if err != nil {
		return err
	}
	cases := []struct {
		name string
		c    candidates
		up   filter.Filters
	}{
		{"Unchanged normal account", base, base.Normal},
		{"Unchanged compact account", base, base.Compact},
		{"Empty account, 850-filter config", base, nil},
		{"Replace every rule on normal account", rewritten, base.Normal},
		{"Partial migration: 212 compact / 213 split rules", base, append(append(filter.Filters{}, base.Compact[:212]...), base.Normal[424:]...)},
		{"Partial migration: 213 compact / 212 split rules", base, append(append(filter.Filters{}, base.Compact[:213]...), base.Normal[426:]...)},
	}
	fmt.Fprintln(w, "\n| Scenario | Normal operations | Compact operations | Hysteresis choice |")
	fmt.Fprintln(w, "|---|---:|---:|---|")
	for _, tc := range cases {
		d := chooseHysteresis(tc.c, tc.up, false)
		fmt.Fprintf(w, "| %s | %d | %d | %s |\n", tc.name, changeCount(tc.up, tc.c.Normal).Total(), changeCount(tc.up, tc.c.Compact).Total(), d.Mode())
	}

	fmt.Fprintln(w, "\n## Repository fixtures")
	fmt.Fprintln(w, "\nEach fixture is evaluated independently, with its current normal output as the installed state. Toggle=true changes and full compaction savings are shown. These small fixtures do not establish how common affected configs are among users.")
	fmt.Fprintln(w, "\n| Fixture | Normal | Compact | Rules combined | Hysteresis changes | Toggle=true changes |")
	fmt.Fprintln(w, "|---|---:|---:|---:|---|---|")
	paths, err := filepath.Glob(filepath.Join(fixtureDir, "*.jsonnet"))
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		return fmt.Errorf("no Jsonnet fixtures found in %s", fixtureDir)
	}
	for _, path := range paths {
		config, err := readConfig(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		c, err := generate(config.Config)
		if err != nil {
			return err
		}
		d := chooseHysteresis(c, c.Normal, false)
		fmt.Fprintf(w, "| %s | %d | %d | %d | %s | %s |\n", filepath.Base(path), len(c.Normal), len(c.Compact), c.CompactedRules, changeCount(c.Normal, d.Filters(c)), changeCount(c.Normal, c.Compact))
	}

	fmt.Fprintln(w, "\n## Shared compactor guard: overlapping rules")
	overlap, err := generate(overlappingRules())
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "\nSix rules combining each of two senders with each of three subjects produce %d distinct normal filters. Independently compacting all six would produce %d distinct filters. The shared guard rejects that expansion and keeps %d filters in either policy. Equal-size candidates are also kept in the normal representation to avoid churn without savings.\n", len(overlap.Normal), overlap.ProposedCompact, len(overlap.Compact))

	fmt.Fprintln(w, "\n## Example explanation before a diff")
	previous, err := generate(synthetic(898, "stable"))
	if err != nil {
		return err
	}
	next, err := generate(synthetic(902, "stable"))
	if err != nil {
		return err
	}
	fmt.Fprintln(w, "\n```text")
	planNotice(w, next, chooseHysteresis(next, previous.Normal, false), previous.Normal, false)
	fmt.Fprintln(w, "```")
	fmt.Fprintln(w, "\nWith the toggle disabled, the same edit keeps normal generation and suggests opting in:")
	fmt.Fprintln(w, "\n```text")
	planNotice(w, next, chooseToggle(false), previous.Normal, false)
	fmt.Fprintln(w, "```")
	fmt.Fprintln(w, "\nNotices are suppressed when there is no filter diff. The grouping explanation is emitted only when the alternate representation would retain installed filters that the selected representation removes.")
	return nil
}
