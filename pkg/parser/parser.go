package parser

import (
	"github.com/pkg/errors"

	cfg "github.com/mbrt/gmailctl/pkg/config/v2alpha1"
)

// Rule is an intermediate representation of a Gmail filter.
type Rule struct {
	Criteria CriteriaAST
	Actions  cfg.Actions
}

// Parse parses config file rules into their intermediate representation.
//
// Note that the number of rules and their contents might be different than the
// original, because symplifications will be performed on the data.
func Parse(config cfg.Config) ([]Rule, error) {
	cmap, err := parseNamedFilters(config.Filters)
	if err != nil {
		return nil, err
	}

	var res []Rule
	for i, rule := range config.Rules {
		crit, err := parseCriteria(rule.Filter, cmap)
		if err != nil {
			return nil, errors.Wrapf(err, "error parsing criteria for rule #d", i)
		}

		simpleCrit, err := SimplifyCriteria(crit)
		if err != nil {
			return nil, errors.Wrapf(err, "error simplifying criteria for rule #d", i)
		}

		// The criteria can be split in multiple ones. In that case we just need
		// to apply the same actions for all of them.
		for _, c := range simpleCrit {
			res = append(res, Rule{
				Criteria: c,
				Actions:  rule.Actions,
			})
		}
	}

	return res, nil
}

// namedCriteriaMap maps a named filter to its parsed representation.
type namedCriteriaMap map[string]CriteriaAST

func parseCriteria(f cfg.FilterNode, nmap namedCriteriaMap) (CriteriaAST, error) {
	// TODO
	return nil, nil
}

func parseNamedFilters(filters []cfg.NamedFilter) (namedCriteriaMap, error) {
	m := namedCriteriaMap{}

	for _, f := range filters {
		c, err := parseCriteria(f.Query, m)
		if err != nil {
			return nil, errors.Wrapf(err, "error parsing filter '%s'", f.Name)
		}
		m[f.Name] = c
	}

	return m, nil
}
