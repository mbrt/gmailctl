package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimplify(t *testing.T) {
	expr := or(
		fn1(FunctionFrom, "a"),
		fn1(FunctionFrom, "b"),
		fn1(FunctionSubject, "c"),
		and(
			fn1(FunctionList, "d"),
			and(
				fn1(FunctionFrom, "e"),
				not(not(
					fn1(FunctionList, "f"),
				)),
			),
		),
	)

	expected := or(
		and(
			fn(FunctionList, OperationAnd, "f", "d"),
			fn(FunctionFrom, OperationAnd, "e"),
		),
		fn(FunctionSubject, OperationOr, "c"),
		fn(FunctionFrom, OperationOr, "a", "b"),
	)
	got, err := SimplifyCriteria(expr)
	assert.Nil(t, err)

	// Maps make the children sorting pseudo-random. We have to sort
	// the trees to be able to find make it deterministic.
	sortTree(expected)
	sortTree(got)
	assert.Equal(t, expected, got)

}

func and(children ...CriteriaAST) *Node {
	return &Node{
		Operation: OperationAnd,
		Children:  children,
	}
}

func or(children ...CriteriaAST) *Node {
	return &Node{
		Operation: OperationOr,
		Children:  children,
	}
}

func not(child CriteriaAST) *Node {
	return &Node{
		Operation: OperationNot,
		Children:  []CriteriaAST{child},
	}
}

func fn(ftype FunctionType, op OperationType, args ...string) *Leaf {
	return &Leaf{
		Function: ftype,
		Grouping: op,
		Args:     args,
	}
}

func fn1(ftype FunctionType, arg string) *Leaf {
	return fn(ftype, OperationNone, arg)
}
