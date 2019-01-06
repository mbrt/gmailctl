package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		Operation: OperationAnd,
		Children:  []CriteriaAST{child},
	}
}

func fn(ftype FunctionType, arg string) *Leaf {
	return &Leaf{
		Function: ftype,
		Grouping: OperationNone,
		Args:     []string{arg},
	}
}

func TestSimplify(t *testing.T) {
	expr := or(
		fn(FunctionFrom, "a"),
		fn(FunctionFrom, "b"),
		fn(FunctionSubject, "c"),
		and(
			fn(FunctionList, "d"),
			and(
				fn(FunctionFrom, "e"),
				not(not(
					fn(FunctionList, "f"),
				)),
			),
		),
	)

	expected := []CriteriaAST{
		and(
			&Leaf{
				Function: FunctionList,
				Grouping: OperationAnd,
				Args:     []string{"d", "f"},
			},
			&Leaf{
				Function: FunctionFrom,
				Grouping: OperationAnd,
				Args:     []string{"e"},
			},
		),
		&Leaf{
			Function: FunctionSubject,
			Grouping: OperationOr,
			Args:     []string{"c"},
		},
		&Leaf{
			Function: FunctionFrom,
			Grouping: OperationOr,
			Args:     []string{"a", "b"},
		},
	}
	got, err := SimplifyCriteria(expr)
	assert.Nil(t, err)

	// Maps make the children sorting pseudo-random. We have to sort
	// the trees to be able to find make it deterministic.
	sortTreeNodes(expected)
	sortTreeNodes(got)
	assert.Equal(t, expected, got)

}
