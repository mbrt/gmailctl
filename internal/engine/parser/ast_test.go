package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestSimplifyKeepsRawArgsSeparate(t *testing.T) {
	// The raw flag applies to all the arguments of a leaf, so merging would
	// make "John Smith" lose its quotes and match "John" OR "Smith".
	expr := or(
		fn1(FunctionFrom, "John Smith"),
		raw(fn1(FunctionFrom, "-(foo bar)")),
		fn1(FunctionFrom, "b"),
		raw(fn1(FunctionFrom, `"c d"`)),
	)
	expected := or(
		fn(FunctionFrom, OperationOr, "John Smith", "b"),
		raw(fn(FunctionFrom, OperationOr, "-(foo bar)", `"c d"`)),
	)

	// Map iteration order is random, so repeat to make sure the result is
	// deterministic.
	for range 20 {
		got, err := SimplifyCriteria(expr.Clone())
		require.Nil(t, err)
		require.Equal(t, expected, got)
	}
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

func raw(l *Leaf) *Leaf {
	l.IsRaw = true
	return l
}
