package v1alpha3

import (
	v2 "github.com/mbrt/gmailctl/pkg/config/v1alpha2"
)

// Import converts a v2 config into a v3.
func Import(cfg v2.Config) (Config, error) {
	var nfs []NamedFilter
	for _, nf := range cfg.Filters {
		nfs = append(nfs, NamedFilter{
			Name:  nf.Name,
			Query: convertFilterNode(nf.Query),
		})
	}

	var rules []Rule
	for _, v2r := range cfg.Rules {
		rules = append(rules, convertRule(v2r))
	}

	return Config{
		Version: Version,
		Author:  cfg.Author,
		Filters: nfs,
		Rules:   rules,
	}, nil
}

func convertRule(r v2.Rule) Rule {
	return Rule{
		Filter:  convertFilterNode(r.Filter),
		Actions: r.Actions,
	}
}

func convertFilterNode(n v2.FilterNode) FilterNode {
	var not *FilterNode
	if n.Not != nil {
		f := convertFilterNode(*n.Not)
		not = &f
	}

	return FilterNode{
		RefName: n.RefName,
		And:     convertFilterNodes(n.And),
		Or:      convertFilterNodes(n.Or),
		Not:     not,
		From:    n.From,
		To:      n.To,
		Cc:      n.Cc,
		Subject: n.Subject,
		List:    n.List,
		Has:     n.Has,
		Query:   n.Query,
	}
}

func convertFilterNodes(ns []v2.FilterNode) []FilterNode {
	var res []FilterNode
	for _, f := range ns {
		res = append(res, convertFilterNode(f))
	}
	return res
}
