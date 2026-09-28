package main

import (
	"fmt"
	"io"

	"github.com/mbrt/gmailctl/internal/engine/filter"
)

type decision struct {
	Compact bool
	Reason  string
}

func (d decision) Mode() string {
	if d.Compact {
		return "compact"
	}
	return "normal"
}

func (d decision) Filters(c candidates) filter.Filters {
	if d.Compact {
		return c.Compact
	}
	return c.Normal
}

func planNotice(w io.Writer, c candidates, d decision, upstream filter.Filters, offline bool) {
	desired := d.Filters(c)
	if offline {
		fmt.Fprintf(w, "Note: exporting %s filters because %s.\n", d.Mode(), d.Reason)
		fmt.Fprintf(w, "Normal generation: %d filters; compact generation: %d filters.\n", len(c.Normal), len(c.Compact))
		return
	}
	delta := changeCount(upstream, desired)
	if delta.Total() == 0 {
		return
	}
	fmt.Fprintf(w, "Note: using %s filters because %s.\n", d.Mode(), d.Reason)
	fmt.Fprintf(w, "Normal generation: %d filters; compact generation: %d filters (%d rules combined).\n", len(c.Normal), len(c.Compact), c.CompactedRules)
	if d.Compact && c.ProposedCompact > len(c.Normal) {
		fmt.Fprintf(w, "Combining these rules would increase the distinct count to %d; retaining normal grouping.\n", c.ProposedCompact)
	}
	if !d.Compact && len(c.Normal) > highThreshold && len(c.Compact) < len(c.Normal) {
		fmt.Fprintf(w, "This config is close to the filter limit. Setting settings.compact to true would reduce it to %d filters.\n", len(c.Compact))
	}
	// Check whether choosing the other representation would preserve any of
	// the installed filters being removed. This is evidence of regrouping,
	// without claiming all changes come from compaction or knowing old intent.
	other := c.Normal
	if !d.Compact {
		other = c.Compact
	}
	opposite, selected := filterSet(other), filterSet(desired)
	regrouped := 0
	for k := range filterSet(upstream) {
		if opposite[k] && !selected[k] {
			regrouped++
		}
	}
	if regrouped > 0 {
		fmt.Fprintf(w, "This grouping choice replaces %d installed filters that the other mode would retain.\n", regrouped)
		fmt.Fprintln(w, "Grouping preserves rule meaning; the diff can include replacements beyond your edits.")
	}
	peak := len(upstream) + delta.Added
	if peak > 1000 {
		fmt.Fprintf(w, "Capacity: the current create-before-delete apply order would peak at %d filters (limit 1000).\n", peak)
	}
}
