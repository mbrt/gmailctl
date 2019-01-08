package filter

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/mbrt/gmailctl/pkg/parser"
)

// FromRules translates rules into entries that map directly into Gmail filters.
func FromRules(rs []parser.Rule) (Filters, error) {
	res := Filters{}
	for i, rule := range rs {
		filters, err := fromRule(rule)
		if err != nil {
			return res, errors.Wrap(err, fmt.Sprintf("error generating rule #%d", i))
		}
		res = append(res, filters...)
	}
	return res, nil
}

func fromRule(rule parser.Rule) ([]Filter, error) {
	criteria := generateCriteria(rule.Criteria)
	actions := generateActions(rule.Actions)
	return combineCriteriaWithActions(criteria, actions), nil
}

func generateCriteria(crit parser.CriteriaAST) Criteria {
	// TODO
	return Criteria{}
}

func generateActions(actions parser.Actions) []Actions {
	res := []Actions{
		{
			Archive:       actions.Archive,
			Delete:        actions.Delete,
			MarkImportant: actions.MarkImportant,
			MarkRead:      actions.MarkRead,
			Category:      actions.Category,
		},
	}

	// Since every action can contain a single lable only, we might need to produce multiple actions
	for i, label := range actions.Labels {
		if i == 0 {
			// The first label can stay in the first action
			res[0].AddLabel = label
		} else {
			// All the subsequent labels need a separate action
			res = append(res, Actions{AddLabel: label})
		}
	}

	return res
}

func combineCriteriaWithActions(criteria Criteria, actions []Actions) Filters {
	// We have to duplicate the criteria for all the given actions
	res := make(Filters, len(actions))
	for i, action := range actions {
		res[i] = Filter{
			Criteria: criteria,
			Action:   action,
		}
	}
	return res
}
