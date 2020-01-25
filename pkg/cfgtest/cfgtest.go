package cfgtest

import (
	"fmt"
	"strings"

	cfg "github.com/mbrt/gmailctl/pkg/config/v1alpha3"
	"github.com/mbrt/gmailctl/pkg/parser"
)

type matchType int

const (
	matchTypeExact = iota
	matchTypeSuffix
	matchTypeContains
)

func Test(rules []parser.Rule, tests []cfg.Test) error {
	return nil
}

func NewEvaluator(criteria parser.CriteriaAST) (RuleEvaluator, error) {
	v := evalBuilder{}
	criteria.AcceptVisitor(&v)
	return v.Res, v.Err
}

type RuleEvaluator interface {
	Match(msg cfg.Message) bool
}

type andNode struct {
	children []RuleEvaluator
}

func (n andNode) Match(msg cfg.Message) bool {
	for _, c := range n.children {
		if !c.Match(msg) {
			return false
		}
	}
	return true
}

type orNode struct {
	children []RuleEvaluator
}

func (n orNode) Match(msg cfg.Message) bool {
	for _, c := range n.children {
		if c.Match(msg) {
			return true
		}
	}
	return false
}

type notNode struct {
	child RuleEvaluator
}

func (n notNode) Match(msg cfg.Message) bool {
	return !n.child.Match(msg)
}

type funcNode struct {
	op        parser.FunctionType
	expected  string
	matchType matchType
}

func (n funcNode) Match(msg cfg.Message) bool {
	var fields []string

	switch n.op {
	case parser.FunctionFrom:
		fields = []string{msg.From}
	case parser.FunctionTo:
		fields = msg.To
	case parser.FunctionCc:
		fields = msg.Cc
	case parser.FunctionBcc:
		fields = msg.Bcc
	case parser.FunctionList:
		fields = msg.Lists
	case parser.FunctionSubject:
		fields = []string{msg.Subject}
	case parser.FunctionHas:
		fields = []string{msg.Body}
	}

	for _, f := range fields {
		isMatch := false

		switch n.matchType {
		case matchTypeExact:
			isMatch = f == n.expected
		case matchTypeSuffix:
			isMatch = strings.HasSuffix(f, n.expected)
		case matchTypeContains:
			isMatch = strings.Contains(f, n.expected)
		}

		// If there's no match, continue searching.
		if isMatch {
			return true
		}
	}

	return false
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
		r.Err = fmt.Errorf("unsupported 'raw' function: %v", n)
		return
	}

	var rules []RuleEvaluator

	switch n.Function {
	case parser.FunctionFrom, parser.FunctionCc, parser.FunctionBcc, parser.FunctionList:
		rules = emailField(n.Function, n.Args)
	case parser.FunctionTo:
		rules = expandTo(n.Args)
	case parser.FunctionSubject:
		rules = freeTextField(n.Function, n.Args)
	case parser.FunctionHas:
		rules = expandHas(n.Args)
	case parser.FunctionQuery:
		r.Err = fmt.Errorf("unsupported unconstrained query: '%v'", n)
		return
	default:
		r.Err = fmt.Errorf("unsupported function: %s", n.Function)
		return
	}

	if len(rules) == 0 {
		r.Err = fmt.Errorf("empty leaf: %v", n)
		return
	}
	if len(rules) == 1 {
		// No need for grouping.
		r.Res = rules[0]
		return
	}

	// Expand into 'and' and 'or' explicitly.
	switch n.Grouping {
	case parser.OperationAnd:
		r.Res = &andNode{rules}
	case parser.OperationOr:
		r.Res = &orNode{rules}
	default:
		r.Err = fmt.Errorf("unsupported grouping %v", n.Grouping)
	}
}

func expandTo(args []string) []RuleEvaluator {
	// In Gmail, 'to' is a shortcut for (to || cc || bcc || list).
	res := emailField(parser.FunctionTo, args)
	res = append(res, emailField(parser.FunctionCc, args)...)
	res = append(res, emailField(parser.FunctionBcc, args)...)
	res = append(res, emailField(parser.FunctionList, args)...)
	return res
}

func expandHas(args []string) []RuleEvaluator {
	// the 'has' operator basically matches every field.
	res := expandTo(args)
	res = append(res, emailField(parser.FunctionFrom, args)...)
	res = append(res, freeTextField(parser.FunctionSubject, args)...)
	return res
}

func emailField(op parser.FunctionType, args []string) []RuleEvaluator {
	var res []RuleEvaluator

	for _, arg := range args {
		// Gmail doesn't distinguish between @ and .
		r := funcNode{
			op:        op,
			expected:  strings.ReplaceAll(arg, "@", "."),
			matchType: matchTypeExact,
		}
		// Asking for *@gmail.com or @gmail.com is the same and means
		// match the suffix.
		if strings.HasPrefix(arg, "*") {
			r.expected = arg[1:]
			r.matchType = matchTypeSuffix
		} else if strings.HasPrefix(arg, ".") {
			r.matchType = matchTypeSuffix
		}
		res = append(res, r)
	}

	return res
}

func freeTextField(op parser.FunctionType, args []string) []RuleEvaluator {
	var res []RuleEvaluator

	for _, arg := range args {
		res = append(res, funcNode{
			op:        op,
			expected:  arg,
			matchType: matchTypeContains,
		})
	}

	return res
}
