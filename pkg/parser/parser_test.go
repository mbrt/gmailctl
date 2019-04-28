package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"

	cfg "github.com/mbrt/gmailctl/pkg/config"
	cfgv1 "github.com/mbrt/gmailctl/pkg/config/v1alpha3"
)

func readConfig(t *testing.T, path string) cfgv1.Config {
	res, err := cfg.ReadFile(path, path)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func boolptr(a bool) *bool {
	return &a
}

func TestParse(t *testing.T) {
	conf := readConfig(t, "testdata/example.jsonnet")
	expected := []Rule{
		{
			Criteria: and(
				fn(FunctionList, OperationOr,
					"list1",
					"list2",
					"list3",
				),
				not(
					fn(FunctionTo, OperationOr,
						"pippo@gmail.com",
						"pippo@hotmail.com",
					),
				),
			),
			Actions: Actions{
				Labels:  []string{"maillist"},
				Archive: true,
			},
		},
		{
			Criteria: and(
				fn(FunctionTo, OperationAnd, "myalias@gmail.com"),
				fn(FunctionList, OperationOr,
					"list1",
					"list2",
					"list3",
				),
			),
			Actions: Actions{MarkImportant: boolptr(true)},
		},
		{
			Criteria: or(
				fn(FunctionFrom, OperationOr,
					"spammer1", "spammer2",
				),
				fn(FunctionTo, OperationOr,
					"pippo+spammy@gmail.com",
				),
				fn(FunctionSubject, OperationOr,
					"spam mail",
				),
				fn(FunctionHas, OperationOr,
					"buy this thing",
					"very important!!!",
				),
			),
			Actions: Actions{Delete: true},
		},
	}
	got, err := Parse(conf)
	assert.Nil(t, err)
	assert.Equal(t, expected, got)
}
