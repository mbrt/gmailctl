package parser

import (
	"strings"

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

	res := []Rule{}
	for i, rule := range config.Rules {
		crit, err := parseCriteria(rule.Filter, cmap)
		if err != nil {
			return nil, errors.Wrapf(err, "error parsing criteria for rule #%d", i)
		}

		scrit, err := SimplifyCriteria(crit)
		if err != nil {
			return nil, errors.Wrapf(err, "error simplifying criteria for rule #%d", i)
		}

		// The criteria can be split in multiple ones. In that case we just need
		// to apply the same actions for all of them.
		for _, c := range scrit {
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
	if err := checkSyntax(f); err != nil {
		return nil, err
	}

	// Since the node is valid, only one function will be present.
	// This means that we can stop checking after the first valid field.
	if f.RefName != "" {
		return parseRefName(f.RefName, nmap)
	}
	if op, children := parseOperation(f); op != OperationNone {
		var astchildren []CriteriaAST
		for _, c := range children {
			astc, err := parseCriteria(c, nmap)
			if err != nil {
				return nil, err
			}
			astchildren = append(astchildren, astc)
		}
		return &Node{
			Operation: op,
			Children:  astchildren,
		}, nil
	}
	if fn, arg := parseFunction(f); fn != FunctionNone {
		return &Leaf{
			Function: fn,
			Grouping: OperationNone,
			Args:     []string{arg},
		}, nil
	}

	return nil, errors.New("empty filter node")
}

func checkSyntax(f cfg.FilterNode) error {
	if fs := f.NonEmptyFields(); len(fs) != 1 {
		if len(fs) == 0 {
			return errors.New("empty filter node")
		}
		return errors.Errorf("multiple fields specified in the same filter node: %s",
			strings.Join(fs, ","))
	}
	return nil
}

func parseRefName(name string, nmap namedCriteriaMap) (CriteriaAST, error) {
	if crit, ok := nmap[name]; ok {
		return crit, nil
	}
	return nil, errors.Errorf("filter name '%s' not found", name)
}

func parseOperation(f cfg.FilterNode) (OperationType, []cfg.FilterNode) {
	if len(f.And) > 0 {
		return OperationAnd, f.And
	}
	if len(f.Or) > 0 {
		return OperationOr, f.Or
	}
	if f.Not != nil {
		return OperationNot, []cfg.FilterNode{*f.Not}
	}
	return OperationNone, nil
}

func parseFunction(f cfg.FilterNode) (FunctionType, string) {
	if f.From != "" {
		return FunctionFrom, f.From
	}
	if f.To != "" {
		return FunctionTo, f.To
	}
	if f.Cc != "" {
		return FunctionCc, f.Cc
	}
	if f.Subject != "" {
		return FunctionSubject, f.Subject
	}
	if f.List != "" {
		return FunctionList, f.List
	}
	if f.Has != "" {
		return FunctionHas, f.Has
	}
	if f.Query != "" {
		// Query and Has are equivalent
		return FunctionHas, f.Query
	}
	return FunctionNone, ""
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
