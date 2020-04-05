package v1alpha2

import (
	"fmt"

	v1 "github.com/mbrt/gmailctl/pkg/config/v1alpha1"
)

// Import converts a v1 config into a v2.
func Import(cfg v1.Config) (Config, error) {
	cfg, err := v1.ResolveConsts(cfg)
	if err != nil {
		return Config{}, fmt.Errorf("resolving consts: %w", err)
	}

	var rules []Rule
	for _, v1r := range cfg.Rules {
		rules = append(rules, convertRule(v1r))
	}

	return Config{
		Version: Version,
		Author:  Author(cfg.Author),
		Rules:   rules,
	}, nil
}

func convertRule(r v1.Rule) Rule {
	// Optional parameters need to be set only if the original action was
	// explicitly set.
	onlyIfSet := func(a bool) *bool {
		if a {
			return &a
		}
		return nil
	}

	return Rule{
		Filter: convertFilter(r.Filters),
		// We need to explicitate the fields because they are not all
		// the same.
		Actions: Actions{
			Archive:       r.Actions.Archive,
			Delete:        r.Actions.Delete,
			MarkImportant: onlyIfSet(r.Actions.MarkImportant),
			MarkRead:      r.Actions.MarkRead,
			Category:      r.Actions.Category,
			Labels:        r.Actions.Labels,
		},
	}
}

func convertFilter(f v1.Filters) FilterNode {
	// All the filters at this level are in 'and'
	res := convertMatchFilter(f.CompositeFilters.MatchFilters)
	if op := convertMatchFilter(f.Not); !op.Empty() {
		res = and(res, FilterNode{
			Not: &op,
		})
	}
	return and(res, FilterNode{Query: f.Query})
}

func convertMatchFilter(f v1.MatchFilters) FilterNode {
	// This filter is an 'and' of operators, where each of them is an 'or'.
	var res FilterNode

	res = and(res, convertOperand(f.From, func(o string) FilterNode { return FilterNode{From: o} }))
	res = and(res, convertOperand(f.To, func(o string) FilterNode { return FilterNode{To: o} }))
	res = and(res, convertOperand(f.Cc, func(o string) FilterNode { return FilterNode{Cc: o} }))
	res = and(res, convertOperand(f.Subject, func(o string) FilterNode { return FilterNode{Subject: o} }))
	res = and(res, convertOperand(f.Has, func(o string) FilterNode { return FilterNode{Has: o} }))
	res = and(res, convertOperand(f.List, func(o string) FilterNode { return FilterNode{List: o} }))

	return res
}

func convertOperand(v1Ops []string, convert func(string) FilterNode) FilterNode {
	// All operands are in 'or' together
	var ops []FilterNode
	for _, e := range v1Ops {
		ops = append(ops, convert(e))
	}
	if len(ops) == 1 {
		// Simplify if there's only one operand
		return ops[0]
	}
	return FilterNode{
		Or: ops,
	}
}

// and returns a new Filter node that represent an 'and' between
// the two given filters.
//
// This function tries to minimize the number of 'and' fields used
// by merging 'and' fields together if possible.
func and(f1, f2 FilterNode) FilterNode {
	if f1.Empty() {
		return f2
	}
	if f2.Empty() {
		return f1
	}

	// If the filter contains an 'and' operation, get rid of it
	// and return all its operands.
	decompose := func(f FilterNode) []FilterNode {
		if len(f.And) > 0 {
			return f.And
		}
		return []FilterNode{f}
	}

	var res FilterNode
	res.And = append(res.And, decompose(f1)...)
	res.And = append(res.And, decompose(f2)...)
	return res
}
