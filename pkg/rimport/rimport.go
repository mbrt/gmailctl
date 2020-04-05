package rimport

import (
	"errors"
	"fmt"
	"strings"

	config "github.com/mbrt/gmailctl/pkg/config/v1alpha3"
	"github.com/mbrt/gmailctl/pkg/filter"
	"github.com/mbrt/gmailctl/pkg/label"
)

// Import converts a list of filters into config rules, best
// effort quality.
func Import(fs filter.Filters, ls label.Labels) (config.Config, error) {
	var rules []config.Rule
	for i, f := range fs {
		r, err := fromFilter(f)
		if err != nil {
			return config.Config{}, fmt.Errorf("importing filter #%d: %w", i, err)
		}
		rules = append(rules, r)
	}

	var labels []config.Label
	for _, l := range ls {
		labels = append(labels, fromLabel(l))
	}

	return config.Config{
		Version: config.Version,
		Author: config.Author{
			Name:  "YOUR NAME HERE (auto imported)",
			Email: "your-email@gmail.com",
		},
		Labels: labels,
		Rules:  rules,
	}, nil
}

func fromLabel(l label.Label) config.Label {
	var color *config.LabelColor
	if l.Color != nil {
		color = &config.LabelColor{
			Background: l.Color.Background,
			Text:       l.Color.Text,
		}
	}
	return config.Label{
		Name:  l.Name,
		Color: color,
	}
}

func fromFilter(f filter.Filter) (config.Rule, error) {
	n, err := fromCriteria(f.Criteria)
	if err != nil {
		return config.Rule{}, err
	}
	a, err := fromActions(f.Action)
	return config.Rule{
		Filter:  n,
		Actions: a,
	}, err
}

func fromCriteria(c filter.Criteria) (config.FilterNode, error) {
	nodes := []config.FilterNode{}
	// Reduce the need for raw nodes as much as we can, by using regular
	// operators when no problematic chars are found.
	//
	// We need raw nodes because filters can already be escaped, so when
	// exporting again we would double escape those strings.
	if c.From != "" {
		n := config.FilterNode{
			From:      c.From,
			IsEscaped: needsEscape(c.From),
		}
		nodes = append(nodes, n)
	}
	if c.To != "" {
		n := config.FilterNode{
			To:        c.To,
			IsEscaped: needsEscape(c.To),
		}
		nodes = append(nodes, n)
	}
	if c.Subject != "" {
		n := config.FilterNode{
			Subject:   c.Subject,
			IsEscaped: needsEscape(c.Subject),
		}
		nodes = append(nodes, n)
	}
	if c.Query != "" {
		n := config.FilterNode{
			Query: c.Query,
			// IsRaw is implicit for query nodes
		}
		nodes = append(nodes, n)
	}

	if len(nodes) == 0 {
		return config.FilterNode{}, errors.New("empty criteria")
	}
	if len(nodes) == 1 {
		return nodes[0], nil
	}
	return config.FilterNode{
		And: nodes,
	}, nil
}

func needsEscape(s string) bool {
	return strings.ContainsAny(s, ` '"`)
}

func fromActions(c filter.Actions) (config.Actions, error) {
	res := config.Actions{
		Category: c.Category,
		Archive:  c.Archive,
		Delete:   c.Delete,
		MarkRead: c.MarkRead,
		Star:     c.Star,
		Forward:  c.Forward,
	}
	if c.AddLabel != "" {
		res.Labels = []string{c.AddLabel}
	}

	var err error
	res.MarkImportant, err = handleTribool(c.MarkImportant, c.MarkNotImportant)
	if err != nil {
		return res, fmt.Errorf("in 'mark important': %w", err)
	}
	if c.MarkNotSpam {
		res.MarkSpam = boolPtr(false)
	}

	return res, nil
}

func handleTribool(isTrue, isFalse bool) (*bool, error) {
	if isTrue && isFalse {
		return nil, errors.New("cannot be both true and false")
	}
	if isTrue || isFalse {
		// They correctly exclude each other, so:
		// - if isTrue: return *true
		// - if isFalse, then return *false (because isTrue = false)
		return &isTrue, nil
	}
	// Neither is specified
	return nil, nil
}

func boolPtr(v bool) *bool {
	return &v
}
