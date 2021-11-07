package api

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/go-multierror"
	gmailv1 "google.golang.org/api/gmail/v1"

	"github.com/mbrt/gmailctl/internal/engine/filter"
	"github.com/mbrt/gmailctl/internal/engine/gmail"
)

var (
	// keep sorted
	knownCriteriaFields = map[string]bool{
		"ExcludeChats":    true,
		"ForceSendFields": true,
		"From":            true,
		"HasAttachment":   true,
		"NegatedQuery":    true,
		"NullFields":      true,
		"Query":           true,
		"Size":            true,
		"SizeComparison":  true,
		"Subject":         true,
		"To":              true,
	}
	// keep sorted
	knownActionFields = map[string]bool{
		"AddLabelIds":     true,
		"ForceSendFields": true,
		"Forward":         true,
		"NullFields":      true,
		"RemoveLabelIds":  true,
	}
	// keep sorted
	unsupportedCriteriaFields = map[string]bool{
		"ExcludeChats":   true,
		"Size":           true,
		"SizeComparison": true,
	}
	// keep sorted
	unsupportedActionFields = map[string]bool{}
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
	if err := checkUnsupportedFields(*action, unsupportedActionFields); err != nil {
		return res, fmt.Errorf("criteria: %w", err)
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
	// We don't ever generate queries that touch certain fields, so supporting them
	// only for the import phase is not worth the effort. Instead update the regular
	// query field with an equivalent expression.
	if criteria == nil {
		return filter.Criteria{}, errors.New("empty criteria")
	}
	if err := checkUnsupportedFields(*criteria, unsupportedCriteriaFields); err != nil {
		return filter.Criteria{}, fmt.Errorf("criteria: %w", err)
	}

	query := appendQuery(nil, criteria.Query)

	// Negated queries:
	// Note that elements in the negated query are by default in OR together, according
	// to GMail behavior.
	if criteria.NegatedQuery != "" {
		query = appendQuery(query, fmt.Sprintf("-{%s}", criteria.NegatedQuery))
	}

	// HasAttachment:
	if criteria.HasAttachment {
		query = appendQuery(query, "has:attachment")
	}

	return filter.Criteria{
		From:    criteria.From,
		To:      criteria.To,
		Subject: criteria.Subject,
		Query:   strings.Join(query, " "),
	}, nil
}

func checkUnsupportedFields(a interface{}, unsupported map[string]bool) error {
	t := reflect.TypeOf(a)
	v := reflect.ValueOf(a)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		name := t.Field(i).Name

		if !unsupported[name] {
			continue
		}
		// Check that the value for unsupported fields is zero.
		if !isDefault(field) {
			return fmt.Errorf("usage of unsupported field %q (value %v)", name, field.Interface())
		}
	}

	return nil
}

func appendQuery(q []string, a string) []string {
	if a == "" {
		return q
	}
	return append(q, a)
}

func isDefault(v reflect.Value) bool {
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
