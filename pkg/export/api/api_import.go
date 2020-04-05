package api

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
	gmailv1 "google.golang.org/api/gmail/v1"

	"github.com/mbrt/gmailctl/pkg/filter"
	"github.com/mbrt/gmailctl/pkg/gmail"
)

// Import exports Gmail filters into Gmail API objects.
//
// If some filter is invalid, the import skips it and returns only the
// valid ones, but records and returns the error in the end.
func Import(filters []*gmailv1.Filter, lmap LabelMap) (filter.Filters, error) {
	res := filter.Filters{}
	var reserr error

	for _, gfilter := range filters {
		impFilter, err := importFilter(gfilter, lmap)
		if err != nil {
			// We don't want to return here, but continue and skip the problematic filter
			err = fmt.Errorf("importing filter %q: %w", gfilter.Id, err)
			reserr = multierror.Append(reserr, err)
		} else {
			res = append(res, impFilter)
		}
	}

	return res, reserr
}

func importFilter(gf *gmailv1.Filter, lmap LabelMap) (filter.Filter, error) {
	action, err := importAction(gf.Action, lmap)
	if err != nil {
		return filter.Filter{}, fmt.Errorf("importing action: %w", err)
	}
	criteria, err := importCriteria(gf.Criteria)
	if err != nil {
		return filter.Filter{}, fmt.Errorf("importing criteria: %w", err)
	}
	return filter.Filter{
		ID:       gf.Id,
		Action:   action,
		Criteria: criteria,
	}, nil
}

func importAction(action *gmailv1.FilterAction, lmap LabelMap) (filter.Actions, error) {
	res := filter.Actions{}
	if action == nil {
		return res, errors.New("empty action")
	}
	if err := importAddLabels(&res, action.AddLabelIds, lmap); err != nil {
		return res, err
	}
	if err := importRemoveLabels(&res, action.RemoveLabelIds); err != nil {
		return res, err
	}
	res.Forward = action.Forward

	if res.Empty() {
		return res, errors.New("empty or unsupported action")
	}
	return res, nil
}

func importAddLabels(res *filter.Actions, addLabelIDs []string, lmap LabelMap) error {
	for _, labelID := range addLabelIDs {
		category := importCategory(labelID)
		if category != "" {
			if res.Category != "" {
				return fmt.Errorf("multiple categories: '%s', '%s'", category, res.Category)
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
				return fmt.Errorf("unknown label ID '%s'", labelID)
			}
			res.AddLabel = labelName
		}
	}
	return nil
}

func importRemoveLabels(res *filter.Actions, removeLabelIDs []string) error {
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
			return fmt.Errorf("unupported label to remove %q", labelID)
		}
	}
	return nil
}

func importCategory(labelID string) gmail.Category {
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

func importCriteria(criteria *gmailv1.FilterCriteria) (filter.Criteria, error) {
	if criteria == nil {
		return filter.Criteria{}, errors.New("empty criteria")
	}
	query := criteria.Query

	// We don't ever generate negated queries, so supporting them only for the import
	// is not worth the effort. Instead update the regular query field with an
	// equivalent expression.
	// Note that elements in the negated query are by default in OR together, according
	// to GMail behavior.
	if criteria.NegatedQuery != "" {
		query = fmt.Sprintf("%s -{%s}", query, criteria.NegatedQuery)
	}

	return filter.Criteria{
		From:    criteria.From,
		To:      criteria.To,
		Subject: criteria.Subject,
		Query:   query,
	}, nil
}
