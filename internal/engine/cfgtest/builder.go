package cfgtest

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mbrt/gmailctl/internal/engine/parser"
)

// NewEvaluator creates a RuleEvaluator starting from a parser criteria.
func NewEvaluator(criteria parser.CriteriaAST) (RuleEvaluator, error) {
	v := evalBuilder{}
	criteria.AcceptVisitor(&v)
	return v.Res, v.Err
}

type evalBuilder struct {
	Res RuleEvaluator
	Err error
}

func (r *evalBuilder) VisitNode(n *parser.Node) {
	var children []RuleEvaluator
	for _, c := range n.Children {
		ce, err := NewEvaluator(c)
		if err != nil {
			r.Err = err
			return
		}
		children = append(children, ce)
	}

	switch n.Operation {
	case parser.OperationAnd:
		r.Res = andNode{children}
	case parser.OperationOr:
		r.Res = orNode{children}
	case parser.OperationNot:
		if len(children) != 1 {
			r.Err = fmt.Errorf("unexpected children size for 'not' node: %d", len(children))
		}
		r.Res = notNode{children[0]}
	default:
		r.Err = fmt.Errorf("unsupported operation %s", n.Operation)
	}
}

func (r *evalBuilder) VisitLeaf(n *parser.Leaf) {
	if n.IsRaw {
		r.Err = fmt.Errorf("unsupported 'raw query': %v", n)
		return
	}

	var rules []RuleEvaluator

	switch n.Function {
	case parser.FunctionFrom, parser.FunctionCc, parser.FunctionBcc, parser.FunctionList, parser.FunctionReplyTo:
		rules = expandAll(n.Args, func(a string) RuleEvaluator {
			return emailField(toMatchField(n.Function), a)
		})
	case parser.FunctionTo:
		rules = expandAll(n.Args, expandTo)
	case parser.FunctionSubject:
		rules = expandAll(n.Args, func(a string) RuleEvaluator {
			return freeTextField(matchFieldSubject, a)
		})
	case parser.FunctionHas:
		rules = expandAll(n.Args, expandHas)
	case parser.FunctionQuery:
		r.Err = fmt.Errorf("unsupported unconstrained query: '%v'", n)
		return
	default:
		r.Err = fmt.Errorf("unsupported function: %s", n.Function)
		return
	}

	r.Res, r.Err = group(n.Grouping, rules)
}

// expandAll applies the given expander to all the arguments.
func expandAll(args []string, f func(arg string) RuleEvaluator) []RuleEvaluator {
	var res []RuleEvaluator
	for _, arg := range args {
		res = append(res, f(arg))
	}
	return res
}

// expandTo expands the 'to' function into the corresponding evaluators.
//
// In Gmail, 'to' is a shortcut for (to || cc || bcc || list).
func expandTo(arg string) RuleEvaluator {
	return orNode{
		[]RuleEvaluator{
			emailField(matchFieldTo, arg),
			emailField(matchFieldCc, arg),
			emailField(matchFieldBcc, arg),
			emailField(matchFieldLists, arg),
		},
	}
}

// expandHas expands the 'has' operator into the corresponding evaluators.
//
// The 'has' operator basically matches every field.
// In input you have a list of items, like "this", "two words", in output evaluators
// that match them in any possible field (to, from, subject, body, ...).
func expandHas(arg string) RuleEvaluator {
	return orNode{
		[]RuleEvaluator{
			expandTo(arg),
			emailField(matchFieldFrom, arg),
			freeTextField(matchFieldSubject, arg),
			freeTextField(matchFieldBody, arg),
		},
	}
}

func emailField(f matchField, arg string) RuleEvaluator {
	// Gmail doesn't distinguish between @ and .
	r := funcNode{
		field:     f,
		expected:  normalizeField(arg),
		matchType: matchTypeExact,
	}
	// Asking for *@gmail.com or @gmail.com is the same and means
	// match the suffix.
	if strings.HasPrefix(r.expected, "*") {
		r.expected = r.expected[1:]
		r.matchType = matchTypeSuffix
	} else if strings.HasPrefix(r.expected, ".") {
		r.matchType = matchTypeSuffix
	}
	return r
}

func freeTextField(f matchField, arg string) RuleEvaluator {
	return funcNode{
		field:     f,
		expected:  normalizeField(arg),
		matchType: matchTypeContains,
	}
}

// group returns an evaluator built by grouping together the given ones with
// an operator.
func group(op parser.OperationType, rs []RuleEvaluator) (RuleEvaluator, error) {
	if len(rs) == 0 {
		return nil, errors.New("empty children, cannot group")
	}
	if op != parser.OperationNot && len(rs) == 1 {
		// No need for grouping.
		return rs[0], nil
	}

	switch op {
	case parser.OperationAnd:
		return andNode{rs}, nil
	case parser.OperationOr:
		return orNode{rs}, nil
	case parser.OperationNot:
		if len(rs) != 1 {
			return nil, fmt.Errorf("unexpected children size for 'not' node: %d", len(rs))
		}
		return notNode{rs[0]}, nil
	default:
		return nil, fmt.Errorf("unsupported operation %s", op)
	}
}

func toMatchField(f parser.FunctionType) matchField {
	switch f {
	case parser.FunctionFrom:
		return matchFieldFrom
	case parser.FunctionTo:
		return matchFieldTo
	case parser.FunctionCc:
		return matchFieldCc
	case parser.FunctionBcc:
		return matchFieldBcc
	case parser.FunctionReplyTo:
		return matchFieldReplyTo
	case parser.FunctionList:
		return matchFieldLists
	case parser.FunctionSubject:
		return matchFieldSubject
	default:
		return matchFieldUnknown
	}
}
