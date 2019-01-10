package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mbrt/gmailctl/pkg/gmail"
	"github.com/mbrt/gmailctl/pkg/parser"
)

func TestSimpleFilter(t *testing.T) {
	rules := []parser.Rule{
		{
			Criteria: &parser.Leaf{
				Function: parser.FunctionFrom,
				Grouping: parser.OperationOr,
				Args:     []string{"a", "b", "with spaces"},
			},
			Actions: parser.Actions{
				Archive: true,
			},
		},
	}
	expected := Filters{
		{
			Criteria: Criteria{
				From: `{a b "with spaces"}`,
			},
			Action: Actions{
				Archive: true,
			},
		},
	}
	got, err := FromRules(rules)
	assert.Nil(t, err)
	assert.Equal(t, expected, got)
}

func TestAndNode(t *testing.T) {
	rules := []parser.Rule{
		{
			Criteria: &parser.Node{
				Operation: parser.OperationAnd,
				Children: []parser.CriteriaAST{
					&parser.Leaf{
						Function: parser.FunctionFrom,
						Args:     []string{"a"},
					},
					&parser.Leaf{
						Function: parser.FunctionTo,
						Grouping: parser.OperationAnd,
						Args:     []string{"a", "b", "c"},
					},
				},
			},
			Actions: parser.Actions{
				Delete:   true,
				Category: gmail.CategoryForums,
			},
		},
	}
	expected := Filters{
		{
			Criteria: Criteria{
				From: "a",
				To:   "(a b c)",
			},
			Action: Actions{
				Delete:   true,
				Category: gmail.CategoryForums,
			},
		},
	}
	got, err := FromRules(rules)
	assert.Nil(t, err)
	assert.Equal(t, expected, got)
}

func TestNotOr(t *testing.T) {
	rules := []parser.Rule{
		{
			Criteria: &parser.Node{
				Operation: parser.OperationNot,
				Children: []parser.CriteriaAST{
					&parser.Node{
						Operation: parser.OperationOr,
						Children: []parser.CriteriaAST{
							&parser.Leaf{
								Function: parser.FunctionTo,
								Grouping: parser.OperationOr,
								Args:     []string{"a", "b"},
							},
							&parser.Leaf{
								Function: parser.FunctionCc,
								Grouping: parser.OperationAnd,
								Args:     []string{"c", "d"},
							},
						},
					},
				},
			},
			Actions: parser.Actions{
				MarkRead: true,
			},
		},
	}
	expected := Filters{
		{
			Criteria: Criteria{
				Query: "-{to:{a b} cc:(c d)}",
			},
			Action: Actions{
				MarkRead: true,
			},
		},
	}
	got, err := FromRules(rules)
	assert.Nil(t, err)
	assert.Equal(t, expected, got)
}

func TestQuoting(t *testing.T) {
	rules := []parser.Rule{
		{
			Criteria: &parser.Node{
				Operation: parser.OperationAnd,
				Children: []parser.CriteriaAST{
					&parser.Leaf{
						Function: parser.FunctionHas,
						Grouping: parser.OperationAnd,
						Args:     []string{"foo", "this is quoted"},
					},
					&parser.Leaf{
						Function: parser.FunctionQuery,
						Args:     []string{`from:foo has:spreadsheet`},
					},
				},
			},
			Actions: parser.Actions{
				MarkImportant: true,
			},
		},
	}
	expected := Filters{
		{
			Criteria: Criteria{
				Query: `(foo "this is quoted") from:foo has:spreadsheet`,
			},
			Action: Actions{
				MarkImportant: true,
			},
		},
	}
	got, err := FromRules(rules)
	assert.Nil(t, err)
	assert.Equal(t, expected, got)
}

func TestSplitActions(t *testing.T) {
	rules := []parser.Rule{
		{
			Criteria: &parser.Leaf{
				Function: parser.FunctionFrom,
				Args:     []string{"a"},
			},
			Actions: parser.Actions{
				Archive:  true,
				MarkRead: true,
				Labels:   []string{"l1", "l2", "l3"},
			},
		},
	}
	expected := Filters{
		{
			Criteria: Criteria{
				From: "a",
			},
			Action: Actions{
				Archive:  true,
				MarkRead: true,
				AddLabel: "l1",
			},
		},
		{
			Criteria: Criteria{
				From: "a",
			},
			Action: Actions{
				AddLabel: "l2",
			},
		},
		{
			Criteria: Criteria{
				From: "a",
			},
			Action: Actions{
				AddLabel: "l3",
			},
		},
	}
	got, err := FromRules(rules)
	assert.Nil(t, err)
	assert.Equal(t, expected, got)
}
