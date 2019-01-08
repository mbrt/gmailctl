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

	expected := []CriteriaAST{
		and(
			fn(FunctionList, OperationAnd, "f", "d"),
			fn(FunctionFrom, OperationAnd, "e"),
		),
		fn(FunctionSubject, OperationOr, "c"),
		fn(FunctionFrom, OperationOr, "a", "b"),
	}
	got, err := SimplifyCriteria(expr)
	assert.Nil(t, err)

	// Maps make the children sorting pseudo-random. We have to sort
	// the trees to be able to find make it deterministic.
	sortTreeNodes(expected)
	sortTreeNodes(got)
	assert.Equal(t, expected, got)

}
