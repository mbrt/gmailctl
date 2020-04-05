package api

import (
	"errors"
	"fmt"

	gmailv1 "google.golang.org/api/gmail/v1"

	"github.com/mbrt/gmailctl/pkg/filter"
	"github.com/mbrt/gmailctl/pkg/gmail"
)

// Export exports Gmail filters into Gmail API objects
func Export(filters filter.Filters, lmap LabelMap) ([]*gmailv1.Filter, error) {
	res := make([]*gmailv1.Filter, len(filters))
	for i, filter := range filters {
		ef, err := export(filter, lmap)
		if err != nil {
			return nil, fmt.Errorf("exporting filter #%d: %w", i, err)
		}
		res[i] = ef
	}
	return res, nil
}

func export(filter filter.Filter, lmap LabelMap) (*gmailv1.Filter, error) {
	if filter.Action.Empty() {
		return nil, errors.New("no action specified")
	}
	if filter.Criteria.Empty() {
		return nil, errors.New("no criteria specified")
	}

	action, err := exportAction(filter.Action, lmap)
	if err != nil {
		return nil, fmt.Errorf("in export action: %w", err)
	}
	criteria, err := exportCriteria(filter.Criteria)
	if err != nil {
		return nil, fmt.Errorf("in export criteria: %w", err)
	}

	return &gmailv1.Filter{
		Action:   action,
		Criteria: criteria,
	}, nil
}

func exportAction(action filter.Actions, lmap LabelMap) (*gmailv1.FilterAction, error) {
	lops := labelOps{}
	exportFlags(action, &lops)

	if action.Category != "" {
		cat, err := exportCategory(action.Category)
		if err != nil {
			return nil, err
		}
		lops.AddLabel(cat)
	}
	if action.AddLabel != "" {
		id, ok := lmap.NameToID(action.AddLabel)
		if !ok {
			return nil, fmt.Errorf("label %q not found", action.AddLabel)
		}
		lops.AddLabel(id)
	}

	return &gmailv1.FilterAction{
		AddLabelIds:    lops.addLabels,
		RemoveLabelIds: lops.removeLabels,
		Forward:        action.Forward,
	}, nil
}

func exportFlags(action filter.Actions, lops *labelOps) {
	if action.Archive {
		lops.RemoveLabel(labelIDInbox)
	}
	if action.Delete {
		lops.AddLabel(labelIDTrash)
	}
	if action.MarkImportant {
		lops.AddLabel(labelIDImportant)
	}
	if action.MarkNotImportant {
		lops.RemoveLabel(labelIDImportant)
	}
	if action.MarkRead {
		lops.RemoveLabel(labelIDUnread)
	}
	if action.MarkNotSpam {
		lops.RemoveLabel(labelIDSpam)
	}
	if action.Star {
		lops.AddLabel(labelIDStar)
	}
}

func exportCategory(category gmail.Category) (string, error) {
	switch category {
	case gmail.CategoryPersonal:
		return labelIDCategoryPersonal, nil
	case gmail.CategorySocial:
		return labelIDCategorySocial, nil
	case gmail.CategoryUpdates:
		return labelIDCategoryUpdates, nil
	case gmail.CategoryForums:
		return labelIDCategoryForums, nil
	case gmail.CategoryPromotions:
		return labelIDCategoryPromotions, nil
	}
	return "", fmt.Errorf("unknown category %q", category)
}

func exportCriteria(criteria filter.Criteria) (*gmailv1.FilterCriteria, error) {
	return &gmailv1.FilterCriteria{
		From:    criteria.From,
		To:      criteria.To,
		Subject: criteria.Subject,
		Query:   criteria.Query,
	}, nil
}

type labelOps struct {
	addLabels    []string
	removeLabels []string
}

func (o *labelOps) AddLabel(name string) {
	o.addLabels = append(o.addLabels, name)
}

func (o *labelOps) RemoveLabel(name string) {
	o.removeLabels = append(o.removeLabels, name)
}
