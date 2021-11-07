package v1alpha3

import (
	"fmt"

	"github.com/hashicorp/go-multierror"

	"github.com/mbrt/gmailctl/cmd/gmailctl-config-migrate/v1alpha2"
	"github.com/mbrt/gmailctl/internal/engine/config/v1alpha3"
)

var dummyFilter = v1alpha3.FilterNode{}

// Import converts a v2 config into a v3.
func Import(cfg v1alpha2.Config) (v1alpha3.Config, error) {
	i := importer{}
	return i.Import(cfg)
}

type importer struct {
	nmap namedFilterMap
	err  error
}

func (i *importer) Import(cfg v1alpha2.Config) (v1alpha3.Config, error) {
	i.importNamedFilters(cfg.Filters)
	finalErr := i.resetError()

	var rules []v1alpha3.Rule
	for _, r := range cfg.Rules {
		rules = append(rules, i.importRule(r))
		if err := i.resetError(); err != nil {
			finalErr = multierror.Append(finalErr,
				fmt.Errorf("in rule %s: %w", r, err),
			)
		}
	}

	return v1alpha3.Config{
		Version: v1alpha3.Version,
		Author: v1alpha3.Author{
			Name:  cfg.Author.Name,
			Email: cfg.Author.Email,
		},
		Rules: rules,
	}, finalErr
}

func (i *importer) importNamedFilters(fs []v1alpha2.NamedFilter) {
	i.nmap = namedFilterMap{}
	var finalErr error

	for _, f := range fs {
		i.nmap[f.Name] = i.importFilter(f.Query)
		if err := i.resetError(); err != nil {
			finalErr = multierror.Append(finalErr,
				fmt.Errorf("in filter '%s' %s: %w", f.Name, f.Query, err),
			)
		}
	}

	i.err = finalErr
}

func (i *importer) importRule(r v1alpha2.Rule) v1alpha3.Rule {
	return v1alpha3.Rule{
		Filter: i.importFilter(r.Filter),
		Actions: v1alpha3.Actions{
			Archive:       r.Actions.Archive,
			Delete:        r.Actions.Delete,
			MarkRead:      r.Actions.MarkRead,
			Star:          r.Actions.Star,
			MarkSpam:      r.Actions.MarkSpam,
			MarkImportant: r.Actions.MarkImportant,
			Category:      r.Actions.Category,
			Labels:        r.Actions.Labels,
		},
	}
}

func (i *importer) importFilter(f v1alpha2.FilterNode) v1alpha3.FilterNode {
	if f.RefName != "" {
		return i.importRefName(f.RefName)
	}

	var not *v1alpha3.FilterNode
	if f.Not != nil {
		nf := i.importFilter(*f.Not)
		not = &nf
	}
	return v1alpha3.FilterNode{
		And:     i.importFilters(f.And),
		Or:      i.importFilters(f.Or),
		Not:     not,
		From:    f.From,
		To:      f.To,
		Cc:      f.Cc,
		Subject: f.Subject,
		List:    f.List,
		Has:     f.Has,
		Query:   f.Query,
	}
}

func (i *importer) importFilters(ns []v1alpha2.FilterNode) []v1alpha3.FilterNode {
	var res []v1alpha3.FilterNode
	for _, f := range ns {
		res = append(res, i.importFilter(f))
	}
	return res
}

func (i *importer) importRefName(name string) v1alpha3.FilterNode {
	if n, ok := i.nmap[name]; ok {
		return n
	}
	i.err = multierror.Append(i.err,
		fmt.Errorf("filter name '%s' not found", name))
	return dummyFilter
}

func (i *importer) resetError() error {
	err := i.err
	i.err = nil
	return err
}

type namedFilterMap map[string]v1alpha3.FilterNode
