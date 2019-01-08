package api

import (
	"fmt"

	"github.com/pkg/errors"
	gmailv1 "google.golang.org/api/gmail/v1"

	"github.com/mbrt/gmailctl/pkg/filter"
	"github.com/mbrt/gmailctl/pkg/gmail"
)

// Exporter exports Gmail filters into Gmail API objects
type Exporter interface {
	// Export exports Gmail filters into Gmail API objects
	Export(filters filter.Filters, lmap LabelMap) ([]*gmailv1.Filter, error)
}

// DefaulExporter returns a default implementation of a Gmail API filter exporter.
func DefaulExporter() Exporter {
	return defaultExporter{}
}

type defaultExporter struct{}

func (de defaultExporter) Export(filters filter.Filters, lmap LabelMap) ([]*gmailv1.Filter, error) {
	res := make([]*gmailv1.Filter, len(filters))
	for i, filter := range filters {
		ef, err := de.export(filter, lmap)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("error exporting filter #%d", i))
		}
		res[i] = ef
	}
	return res, nil
}

func (de defaultExporter) export(filter filter.Filter, lmap LabelMap) (*gmailv1.Filter, error) {
	if filter.Action.Empty() {
		return nil, errors.New("no action specified")
	}
	if filter.Criteria.Empty() {
		return nil, errors.New("no criteria specified")
	}

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

func (de defaultExporter) exportAction(action filter.Actions, lmap LabelMap) (*gmailv1.FilterAction, error) {
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
			return nil, errors.Errorf("label '%s' not found", action.AddLabel)
		}
		addlabels = append(addlabels, id)
	}

	return &gmailv1.FilterAction{
		AddLabelIds:    addlabels,
		RemoveLabelIds: removelabels,
	}, nil
}

func (de defaultExporter) exportCategory(category gmail.Category) (string, error) {
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
	return "", errors.Errorf("unknown category '%s'", category)
}

func (de defaultExporter) exportCriteria(criteria filter.Criteria) (*gmailv1.FilterCriteria, error) {
	return &gmailv1.FilterCriteria{
		From:    criteria.From,
		To:      criteria.To,
		Subject: criteria.Subject,
		Query:   criteria.Query,
	}, nil
}
