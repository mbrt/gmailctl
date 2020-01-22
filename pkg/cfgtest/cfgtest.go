package cfgtest

import (
	"fmt"
	"regexp"

	cfg "github.com/mbrt/gmailctl/pkg/config/v1alpha3"
	"github.com/mbrt/gmailctl/pkg/parser"
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
	op parser.FunctionType
	re *regexp.Regexp
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
		if n.re.MatchString(f) {
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
	// TODO
}
