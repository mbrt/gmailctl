package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mbrt/gmailctl/internal/engine/gmail"
	"github.com/mbrt/gmailctl/internal/engine/parser"
)

func boolptr(a bool) *bool {
	return &a
}

func TestQuotes(t *testing.T) {
	rules := []parser.Rule{
		{
			Criteria: &parser.Leaf{
				Function: parser.FunctionFrom,
				Grouping: parser.OperationOr,
				Args: []string{
					"a",
					"with spaces",
					"with+plus",
					"with+plus@email.com",
					`"already-quoted"`,
				},
			},
			Actions: parser.Actions{
				Archive: true,
			},
		},
	}
	// The plus sign is quoted, except for when in full email addresses.
	expected := Filters{
		{
			Criteria: Criteria{
				From: `{a "with spaces" "with+plus" with+plus@email.com "already-quoted"}`,
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
				MarkImportant: boolptr(true),
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

func TestSplitLeaf(t *testing.T) {
	rule := parser.Rule{
		Criteria: &parser.Leaf{
			Function: parser.FunctionFrom,
			Grouping: parser.OperationOr,
			Args:     []string{"a", "b", "c"},
		},
		Actions: parser.Actions{Archive: true},
	}
	expected := Filters{
		{
			Criteria: Criteria{From: "{a b}"},
			Action:   Actions{Archive: true},
		},
		{
			Criteria: Criteria{From: "c"},
			Action:   Actions{Archive: true},
		},
	}
	got, err := FromRule(rule, 2)
	assert.Nil(t, err)
	assert.Equal(t, expected, got)
}

func TestSplitFail(t *testing.T) {
	rule := parser.Rule{
		Criteria: &parser.Node{
			Operation: parser.OperationAnd,
			Children: []parser.CriteriaAST{
				&parser.Leaf{
					Function: parser.FunctionFrom,
					Grouping: parser.OperationAnd,
					Args:     []string{"d", "e"},
				},
				&parser.Leaf{
					Function: parser.FunctionList,
					Grouping: parser.OperationAnd,
					Args:     []string{"a", "b"},
				},
				&parser.Node{
					Operation: parser.OperationNot,
					Children: []parser.CriteriaAST{
						&parser.Leaf{
							Function: parser.FunctionTo,
							Args:     []string{"e"},
						},
					},
				},
			},
		},
		Actions: parser.Actions{Archive: true},
	}
	expected := Filters{
		{
			Criteria: Criteria{
				From:  "(d e)",
				Query: "list:(a b) -to:e",
			},
			Action: Actions{Archive: true},
		},
	}
	got, err := FromRule(rule, 3)
	assert.Nil(t, err)
	assert.Equal(t, expected, got)
}

func TestSplitComplex(t *testing.T) {
	rule := parser.Rule{
		Criteria: &parser.Node{
			Operation: parser.OperationAnd,
			Children: []parser.CriteriaAST{
				&parser.Leaf{
					Function: parser.FunctionFrom,
					Grouping: parser.OperationOr,
					Args:     []string{"d", "e"},
				},
				&parser.Leaf{
					Function: parser.FunctionList,
					Grouping: parser.OperationOr,
					Args:     []string{"a", "b", "c"},
				},
				&parser.Node{
					Operation: parser.OperationNot,
					Children: []parser.CriteriaAST{
						&parser.Leaf{
							Function: parser.FunctionTo,
							Args:     []string{"e"},
						},
					},
				},
			},
		},
		Actions: parser.Actions{Archive: true},
	}
	expected := Filters{
		{
			Criteria: Criteria{
				From:  "{d e}",
				Query: "list:{a b} -to:e",
			},
			Action: Actions{Archive: true},
		},
		{
			Criteria: Criteria{
				From:  "{d e}",
				Query: "list:c -to:e",
			},
			Action: Actions{Archive: true},
		},
	}
	got, err := FromRule(rule, 7)
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

func TestActions(t *testing.T) {
	rules := []parser.Rule{
		{
			Criteria: &parser.Leaf{
				Function: parser.FunctionFrom,
				Args:     []string{"a"},
			},
			Actions: parser.Actions{
				Archive:       true,
				Delete:        true,
				MarkRead:      true,
				Star:          true,
				MarkSpam:      boolptr(false),
				MarkImportant: boolptr(true),
				Category:      gmail.CategoryForums,
				Forward:       "foo@bar.com",
			},
		},
	}
	expected := Filters{
		{
			Criteria: Criteria{
				From: "a",
			},
			Action: Actions{
				Archive:       true,
				Delete:        true,
				MarkRead:      true,
				Star:          true,
				MarkNotSpam:   true,
				MarkImportant: true,
				Category:      gmail.CategoryForums,
				Forward:       "foo@bar.com",
			},
		},
	}
	got, err := FromRules(rules)
	assert.Nil(t, err)
	assert.Equal(t, expected, got)
}
