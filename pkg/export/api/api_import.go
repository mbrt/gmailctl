package api

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	gmailv1 "google.golang.org/api/gmail/v1"

	"github.com/mbrt/gmailctl/pkg/config"
	"github.com/mbrt/gmailctl/pkg/filter"
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

func (di defaultImporter) importAction(action *gmailv1.FilterAction, lmap LabelMap) (filter.Action, error) {
	res := filter.Action{}
	if err := di.importAddLabels(&res, action.AddLabelIds, lmap); err != nil {
		return res, err
	}
	err := di.importRemoveLabels(&res, action.RemoveLabelIds)
	return res, err
}

func (di defaultImporter) importAddLabels(res *filter.Action, addLabelIDs []string, lmap LabelMap) error {
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

func (di defaultImporter) importRemoveLabels(res *filter.Action, removeLabelIDs []string) error {
	for _, labelID := range removeLabelIDs {
		switch labelID {
		case labelIDInbox:
			res.Archive = true
		case labelIDUnread:
			res.MarkRead = true
		default:
			// filters not added by us are not supported
			return errors.Errorf("unupported label to remove '%s'", labelID)
		}
	}
	return nil
}

func (di defaultImporter) importCategory(labelID string) config.Category {
	switch labelID {
	case labelIDCategoryPersonal:
		return config.CategoryPersonal
	case labelIDCategorySocial:
		return config.CategorySocial
	case labelIDCategoryUpdates:
		return config.CategoryUpdates
	case labelIDCategoryForums:
		return config.CategoryForums
	case labelIDCategoryPromotions:
		return config.CategoryPromotions
	default:
		return ""
	}
}

func (di defaultImporter) importCriteria(action *gmailv1.FilterCriteria) (filter.Criteria, error) {
	return filter.Criteria{
		From:    action.From,
		To:      action.To,
		Subject: action.Subject,
		Query:   action.Query,
	}, nil
}
