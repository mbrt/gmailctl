package cfgtest

import (
	cfg "github.com/mbrt/gmailctl/pkg/config/v1alpha3"
	"github.com/mbrt/gmailctl/pkg/parser"
)

func Test(rules []parser.Rule, tests []cfg.Test) error {
	return nil
}

func match(msg cfg.Message, c parser.CriteriaAST) (bool, error) {
	eval := ruleEvaluator{msg: msg}
	c.AcceptVisitor(&eval)
	return eval.Match, eval.Err
}

type ruleEvaluator struct {
	msg   cfg.Message
	Match bool
	Err   error
}

func (r *ruleEvaluator) VisitNode(n *parser.Node) {
	if n.Operation == parser.OperationNot {
		m, err := match(r.msg, n.Children[0])
		r.Match, r.Err = !m, err
		return
	}

	r.Match = accumulateInit(n.Operation)
	for _, child := range n.Children {
		m, err := match(r.msg, child)
		r.Match = accumulate(n.Operation, r.Match, m)
		if r.Err == nil && err != nil {
			// Never override the first error.
			r.Err = err
		}
	}
}

func (r *ruleEvaluator) VisitLeaf(n *parser.Leaf) {
	// TODO
}

func accumulateInit(op parser.OperationType) bool {
	// We need to start with a true only when we need to accumulate with
	// AND. OR has to start with false and we don't care about NOT.
	return op == parser.OperationAnd
}

func accumulate(op parser.OperationType, init, new bool) bool {
	if op == parser.OperationAnd {
		return init && new
	}
	return init || new
}
