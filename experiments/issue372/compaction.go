// This executable is an isolated design experiment, not a production feature.
package main

import (
	"fmt"

	cfg "github.com/mbrt/gmailctl/internal/engine/config/v1alpha3"
	"github.com/mbrt/gmailctl/internal/engine/filter"
	"github.com/mbrt/gmailctl/internal/engine/parser"
)

// Match the existing compiler's complexity budget. The byte budget is an
// experimental guard, not a documented guarantee about Gmail's query engine.
const compactNodeLimit = 20
const compactByteLimit = 1000

type candidates struct {
	Normal, Compact filter.Filters
	CompactedRules  int
	ProposedCompact int
}

func generate(config cfg.Config) (candidates, error) {
	rules, err := parser.Parse(config)
	if err != nil {
		return candidates{}, err
	}
	var out candidates
	for i, rule := range rules {
		// Reuse the real compiler for normal splitting and action expansion.
		normal, err := filter.FromRules([]parser.Rule{rule})
		if err != nil {
			return out, err
		}
		out.Normal = append(out.Normal, normal...)
		compact, changed, err := compactRule(rule, normal)
		if err != nil {
			return out, fmt.Errorf("rule %d: %w", i, err)
		}
		out.Compact = append(out.Compact, compact...)
		if changed {
			out.CompactedRules++
		}
	}
	out.Normal = unique(out.Normal)
	out.Compact = unique(out.Compact)
	out.ProposedCompact = len(out.Compact)
	// Different rules can share identical split filters. Combining each rule
	// independently can lose that sharing, so count the complete candidates.
	// Avoid a representation migration when it saves no filters.
	if len(out.Compact) >= len(out.Normal) {
		out.Compact = out.Normal
		out.CompactedRules = 0
	}
	return out, nil
}

func compactRule(rule parser.Rule, normal filter.Filters) (filter.Filters, bool, error) {
	node, ok := rule.Criteria.(*parser.Node)
	if !ok || node.Operation != parser.OperationOr {
		return normal, false, nil
	}
	count, structured := complexity(node)
	if !structured || count >= compactNodeLimit {
		return normal, false, nil
	}
	criteria, err := filter.GenerateCriteria(node)
	if err != nil {
		return nil, false, err
	}
	if len(criteria.Query) > compactByteLimit {
		return normal, false, nil
	}

	// A single combined criterion still needs one filter per expanded action.
	// Deduplicate action bundles to avoid repeating them for each old OR branch.
	seen := map[filter.Actions]bool{}
	var out filter.Filters
	for _, f := range normal {
		if !seen[f.Action] {
			seen[f.Action] = true
			out = append(out, filter.Filter{Criteria: criteria, Action: f.Action})
		}
	}
	return out, true, nil
}

func complexity(tree parser.CriteriaAST) (int, bool) {
	switch n := tree.(type) {
	case *parser.Leaf:
		return len(n.Args), !n.IsRaw && n.Function != parser.FunctionQuery
	case *parser.Node:
		count := 1
		for _, child := range n.Children {
			size, structured := complexity(child)
			if !structured {
				return 0, false
			}
			count += size
		}
		return count, true
	default:
		return 0, false
	}
}

// The existing diff compares exactly these values, excluding the server ID.
// Comparable Go structs let the experiment do that without hashing strings or
// invoking the expensive presentation matcher for every candidate.
type filterKey struct {
	Criteria filter.Criteria
	Action   filter.Actions
}

func key(f filter.Filter) filterKey { return filterKey{f.Criteria, f.Action} }

func filterSet(fs filter.Filters) map[filterKey]bool {
	set := make(map[filterKey]bool, len(fs))
	for _, f := range fs {
		set[key(f)] = true
	}
	return set
}

func unique(fs filter.Filters) filter.Filters {
	seen := map[filterKey]bool{}
	out := filter.Filters{}
	for _, f := range fs {
		if !seen[key(f)] {
			seen[key(f)] = true
			out = append(out, f)
		}
	}
	return out
}

type changes struct{ Added, Removed int }

func (c changes) Total() int { return c.Added + c.Removed }

func (c changes) String() string { return fmt.Sprintf("+%d / -%d", c.Added, c.Removed) }

func changeCount(upstream, desired filter.Filters) changes {
	u, d := filterSet(upstream), filterSet(desired)
	var c changes
	for k := range d {
		if !u[k] {
			c.Added++
		}
	}
	for k := range u {
		if !d[k] {
			c.Removed++
		}
	}
	return c
}
