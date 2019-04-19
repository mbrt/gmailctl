package api

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	gmailv1 "google.golang.org/api/gmail/v1"

	"github.com/mbrt/gmailctl/pkg/filter"
	"github.com/mbrt/gmailctl/pkg/gmail"
)

// Importer imports Gmail API objects into filters
type Importer interface {
	// Import exports Gmail filters into Gmail API objects.
	//
	// If some filter is invalid, the import skips it and returns only the
	// valid ones, but records and returns the error in the end.
	Import(filters []*gmailv1.Filter, lmap LabelMap) (filter.Filters, error)
}

// DefaulImporter returns a default implementation of a Gmail API filter exporter.
func DefaulImporter() Importer {
	return defaultImporter{}
}

type defaultImporter struct{}

func (di defaultImporter) Import(filters []*gmailv1.Filter, lmap LabelMap) (filter.Filters, error) {
	res := filter.Filters{}
	var reserr error

	for _, gfilter := range filters {
		impFilter, err := di.importFilter(gfilter, lmap)
		if err != nil {
			// We don't want to return here, but continue and skip the problematic filter
			err = errors.Wrap(err, fmt.Sprintf("error importing filter '%s'", gfilter.Id))
			reserr = multierror.Append(reserr, err)
		} else {
			res = append(res, impFilter)
		}
	}

	return res, reserr
}

func (di defaultImporter) importFilter(gf *gmailv1.Filter, lmap LabelMap) (filter.Filter, error) {
	action, err := di.importAction(gf.Action, lmap)
	if err != nil {
		return filter.Filter{}, errors.Wrap(err, "error importing action")
	}
	criteria, err := di.importCriteria(gf.Criteria)
	if err != nil {
		return filter.Filter{}, errors.Wrap(err, "error importing criteria")
	}
	return filter.Filter{
		ID:       gf.Id,
		Action:   action,
		Criteria: criteria,
	}, nil
}

func (di defaultImporter) importAction(action *gmailv1.FilterAction, lmap LabelMap) (filter.Actions, error) {
	res := filter.Actions{}
	if action == nil {
		return res, errors.New("empty action")
	}
	if err := di.importAddLabels(&res, action.AddLabelIds, lmap); err != nil {
		return res, err
	}
	err := di.importRemoveLabels(&res, action.RemoveLabelIds)

	if res.Empty() {
		return res, errors.New("empty or unsupported action")
	}
	return res, err
}

func (di defaultImporter) importAddLabels(res *filter.Actions, addLabelIDs []string, lmap LabelMap) error {
	for _, labelID := range addLabelIDs {
		category := di.importCategory(labelID)
		if category != "" {
			if res.Category != "" {
				return errors.Errorf("multiple categories: '%s', '%s'", category, res.Category)
			}
			res.Category = category
			continue
		}

		switch labelID {
		case labelIDTrash:
			res.Delete = true
		case labelIDImportant:
			res.MarkImportant = true
		case labelIDStar:
			res.Star = true
		default:
			// it should be a label to add
			labelName, ok := lmap.IDToName(labelID)
			if !ok {
				return errors.Errorf("unknown label ID '%s'", labelID)
			}
			res.AddLabel = labelName
		}
	}
	return nil
}

func (di defaultImporter) importRemoveLabels(res *filter.Actions, removeLabelIDs []string) error {
	for _, labelID := range removeLabelIDs {
		switch labelID {
		case labelIDInbox:
			res.Archive = true
		case labelIDUnread:
			res.MarkRead = true
		case labelIDImportant:
			res.MarkNotImportant = true
		case labelIDSpam:
			res.MarkNotSpam = true
		default:
			// filters not added by us are not supported
			return errors.Errorf("unupported label to remove '%s'", labelID)
		}
	}
	return nil
}

func (di defaultImporter) importCategory(labelID string) gmail.Category {
	switch labelID {
	case labelIDCategoryPersonal:
		return gmail.CategoryPersonal
	case labelIDCategorySocial:
		return gmail.CategorySocial
	case labelIDCategoryUpdates:
		return gmail.CategoryUpdates
	case labelIDCategoryForums:
		return gmail.CategoryForums
	case labelIDCategoryPromotions:
		return gmail.CategoryPromotions
	default:
		return ""
	}
}

func (di defaultImporter) importCriteria(criteria *gmailv1.FilterCriteria) (filter.Criteria, error) {
	if criteria == nil {
		return filter.Criteria{}, errors.New("empty criteria")
	}
	return filter.Criteria{
		From:    criteria.From,
		To:      criteria.To,
		Subject: criteria.Subject,
		Query:   criteria.Query,
	}, nil
}
