package api

import (
	"fmt"

	"github.com/pkg/errors"
	gmailv1 "google.golang.org/api/gmail/v1"

	"github.com/mbrt/gmailfilter/pkg/config"
	"github.com/mbrt/gmailfilter/pkg/filter"
)

const (
	labelIDInbox     = "INBOX"
	labelIDTrash     = "TRASH"
	labelIDImportant = "IMPORTANT"
	labelIDUnread    = "UNREAD"

	labelIDCategoryPersonal   = "CATEGORY_PERSONAL"
	labelIDCategorySocial     = "CATEGORY_SOCIAL"
	labelIDCategoryUpdates    = "CATEGORY_UPDATES"
	labelIDCategoryForums     = "CATEGORY_FORUMS"
	labelIDCategoryPromotions = "CATEGORY_PROMOTIONS"
)

// Exporter exports Gmail filters into Gmail API objects
type Exporter interface {
	// Export exports Gmail filters into Gmail API objects
	Export(filters filter.Filters, lmap LabelMap) ([]gmailv1.Filter, error)
}

// LabelMap maps label names and IDs together.
type LabelMap interface {
	NameToID(name string) (string, bool)
	IDToName(id string) (string, bool)
}

// DefaulExporter returns a default implementation of a Gmail API filter exporter.
func DefaulExporter() Exporter {
	return defaultExporter{}
}

type defaultExporter struct{}

func (de defaultExporter) Export(filters filter.Filters, lmap LabelMap) ([]gmailv1.Filter, error) {
	res := make([]gmailv1.Filter, len(filters))
	for i, filter := range filters {
		ef, err := de.export(filter, lmap)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("error exporting filter #%d", i))
		}
		res[i] = *ef
	}
	return res, nil
}

func (de defaultExporter) export(filter filter.Filter, lmap LabelMap) (*gmailv1.Filter, error) {
	action, err := de.exportAction(filter.Action, lmap)
	if err != nil {
		return nil, errors.Wrap(err, "error in export action")
	}
	criteria, err := de.exportCriteria(filter.Criteria)
	if err != nil {
		return nil, errors.Wrap(err, "error in export criteria")
	}
	return &gmailv1.Filter{
		Action:   action,
		Criteria: criteria,
	}, nil
}

func (de defaultExporter) exportAction(action filter.Action, lmap LabelMap) (*gmailv1.FilterAction, error) {
	addlabels := []string{}
	removelabels := []string{}

	if action.Archive {
		removelabels = append(removelabels, labelIDInbox)
	}
	if action.Delete {
		addlabels = append(addlabels, labelIDTrash)
	}
	if action.MarkImportant {
		addlabels = append(addlabels, labelIDImportant)
	}
	if action.MarkRead {
		removelabels = append(removelabels, labelIDUnread)
	}
	if action.Category != "" {
		cat, err := de.exportCategory(action.Category)
		if err != nil {
			return nil, err
		}
		addlabels = append(addlabels, cat)
	}
	if action.AddLabel != "" {
		id, ok := lmap.NameToID(action.AddLabel)
		if !ok {
			return nil, fmt.Errorf("label '%s' not found", action.AddLabel)
		}
		addlabels = append(addlabels, id)
	}

	return &gmailv1.FilterAction{
		AddLabelIds:    addlabels,
		RemoveLabelIds: removelabels,
	}, nil
}

func (de defaultExporter) exportCategory(category config.Category) (string, error) {
	switch category {
	case config.CategoryPersonal:
		return labelIDCategoryPersonal, nil
	case config.CategorySocial:
		return labelIDCategorySocial, nil
	case config.CategoryUpdates:
		return labelIDCategoryUpdates, nil
	case config.CategoryForums:
		return labelIDCategoryForums, nil
	case config.CategoryPromotions:
		return labelIDCategoryPromotions, nil
	}
	return "", fmt.Errorf("unknown category '%s'", category)
}

func (de defaultExporter) exportCriteria(criteria filter.Criteria) (*gmailv1.FilterCriteria, error) {
	return &gmailv1.FilterCriteria{
		From:    criteria.From,
		To:      criteria.To,
		Subject: criteria.Subject,
		Query:   criteria.Query,
	}, nil
}
