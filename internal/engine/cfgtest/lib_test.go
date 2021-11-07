package cfgtest

import "github.com/mbrt/gmailctl/internal/engine/parser"

func and(children ...parser.CriteriaAST) *parser.Node {
	return &parser.Node{
		Operation: parser.OperationAnd,
		Children:  children,
	}
}

func or(children ...parser.CriteriaAST) *parser.Node {
	return &parser.Node{
		Operation: parser.OperationOr,
		Children:  children,
	}
}

func fn(ftype parser.FunctionType, op parser.OperationType, args ...string) *parser.Leaf {
	return &parser.Leaf{
		Function: ftype,
		Grouping: op,
		Args:     args,
	}
}

func fn1(ftype parser.FunctionType, arg string) *parser.Leaf {
	return fn(ftype, parser.OperationNone, arg)
}
