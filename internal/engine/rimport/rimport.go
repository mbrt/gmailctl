package rimport

import (
	"fmt"
	"strings"

	"github.com/mbrt/gmailctl/internal/engine/config/v1alpha3"
	"github.com/mbrt/gmailctl/internal/engine/filter"
	"github.com/mbrt/gmailctl/internal/engine/label"
	"github.com/mbrt/gmailctl/internal/errors"
	"github.com/mbrt/gmailctl/internal/reporting"
)

// Import converts a list of filters into config rules, best
// effort quality.
func Import(fs filter.Filters, ls label.Labels) (v1alpha3.Config, error) {
	var rules []v1alpha3.Rule
	for i, f := range fs {
		r, err := fromFilter(f)
		if err != nil {
			return v1alpha3.Config{}, errors.WithDetails(
				fmt.Errorf("importing filter #%d: %w", i, err),
				fmt.Sprintf("Filter (internal representation): %s", reporting.Prettify(f, false)))
		}
		rules = append(rules, r)
	}

	var labels []v1alpha3.Label
	for _, l := range ls {
		labels = append(labels, fromLabel(l))
	}

	return v1alpha3.Config{
		Version: v1alpha3.Version,
		Author: v1alpha3.Author{
			Name:  "YOUR NAME HERE (auto imported)",
			Email: "your-email@gmail.com",
		},
		Labels: labels,
		Rules:  rules,
	}, nil
}

func fromLabel(l label.Label) v1alpha3.Label {
	var color *v1alpha3.LabelColor
	if l.Color != nil {
		color = &v1alpha3.LabelColor{
			Background: l.Color.Background,
			Text:       l.Color.Text,
		}
	}
	return v1alpha3.Label{
		Name:  l.Name,
		Color: color,
	}
}

func fromFilter(f filter.Filter) (v1alpha3.Rule, error) {
	n, err := fromCriteria(f.Criteria)
	if err != nil {
		return v1alpha3.Rule{}, err
	}
	a, err := fromActions(f.Action)
	return v1alpha3.Rule{
		Filter:  n,
		Actions: a,
	}, err
}

func fromCriteria(c filter.Criteria) (v1alpha3.FilterNode, error) {
	nodes := []v1alpha3.FilterNode{}
	// Reduce the need for raw nodes as much as we can, by using regular
	// operators when no problematic chars are found.
	//
	// We need raw nodes because filters can already be escaped, so when
	// exporting again we would double escape those strings.
	if c.From != "" {
		n := v1alpha3.FilterNode{
			From:      c.From,
			IsEscaped: needsEscape(c.From),
		}
		nodes = append(nodes, n)
	}
	if c.To != "" {
		n := v1alpha3.FilterNode{
			To:        c.To,
			IsEscaped: needsEscape(c.To),
		}
		nodes = append(nodes, n)
	}
	if c.Subject != "" {
		n := v1alpha3.FilterNode{
			Subject:   c.Subject,
			IsEscaped: needsEscape(c.Subject),
		}
		nodes = append(nodes, n)
	}
	if c.Query != "" {
		n := v1alpha3.FilterNode{
			Query: c.Query,
			// IsRaw is implicit for query nodes
		}
		nodes = append(nodes, n)
	}

	if len(nodes) == 0 {
		return v1alpha3.FilterNode{}, errors.New("empty criteria")
	}
	if len(nodes) == 1 {
		return nodes[0], nil
	}
	return v1alpha3.FilterNode{
		And: nodes,
	}, nil
}

func needsEscape(s string) bool {
	return strings.ContainsAny(s, ` '"`)
}

func fromActions(c filter.Actions) (v1alpha3.Actions, error) {
	res := v1alpha3.Actions{
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
