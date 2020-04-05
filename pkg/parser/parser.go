package parser

import (
	"fmt"
	"errors"
	"strings"

	cfg "github.com/mbrt/gmailctl/pkg/config/v1alpha3"
)

// Rule is an intermediate representation of a Gmail filter.
type Rule struct {
	Criteria CriteriaAST
	Actions  Actions
}

// Actions contains the actions to be applied to a set of emails.
type Actions cfg.Actions

// Parse parses config file rules into their intermediate representation.
//
// Note that the number of rules and their contents might be different than the
// original, because symplifications will be performed on the data.
func Parse(config cfg.Config) ([]Rule, error) {
	res := []Rule{}
	for i, rule := range config.Rules {
		crit, err := parseCriteria(rule.Filter)
		if err != nil {
			return nil, fmt.Errorf("parsing criteria for rule #%d: %w", i, err)
		}

		scrit, err := SimplifyCriteria(crit)
		if err != nil {
			return nil, fmt.Errorf("simplifying criteria for rule #%d: %w", i, err)
		}

		res = append(res, Rule{
			Criteria: scrit,
			Actions:  Actions(rule.Actions),
		})
	}

	return res, nil
}

func parseCriteria(f cfg.FilterNode) (CriteriaAST, error) {
	if err := checkSyntax(f); err != nil {
		return nil, err
	}

	// Since the node is valid, only one function will be present.
	// This means that we can stop checking after the first valid field.
	if op, children := parseOperation(f); op != OperationNone {
		var astchildren []CriteriaAST
		for _, c := range children {
			astc, err := parseCriteria(c)
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
			IsRaw:    f.IsEscaped,
		}, nil
	}

	return nil, errors.New("empty filter node")
}

func checkSyntax(f cfg.FilterNode) error {
	fs := f.NonEmptyFields()
	if len(fs) != 1 {
		if len(fs) == 0 {
			return errors.New("empty filter node")
		}
		return fmt.Errorf("multiple fields specified in the same filter node: %s",
			strings.Join(fs, ","))
	}
	if !f.IsEscaped {
		return nil
	}

	// Check that 'isRaw' is used correctly
	allowed := []string{"from", "to", "subject"}
	for _, s := range allowed {
		if fs[0] == s {
			return nil
		}
	}
	return fmt.Errorf("'isRaw' can be used only with fields %s", strings.Join(allowed, ", "))
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
	if f.Bcc != "" {
		return FunctionBcc, f.Bcc
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
		return FunctionQuery, f.Query
	}
	return FunctionNone, ""
}
