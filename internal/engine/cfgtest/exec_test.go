package cfgtest

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mbrt/gmailctl/internal/engine/apply"
	"github.com/mbrt/gmailctl/internal/engine/config"
	"github.com/mbrt/gmailctl/internal/engine/config/v1alpha3"
)

func readConfig(t *testing.T, path string) v1alpha3.Config {
	t.Helper()
	path = filepath.Join("testdata", path)
	res, err := config.ReadFile(path, path)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestExec(t *testing.T) {
	tests := []struct {
		path        string
		numErrs     int
		expectedOut string
	}{
		{path: "pass.jsonnet"},
		{
			path:    "invalid.jsonnet",
			numErrs: 1,
			expectedOut: `Failed: 1/1

Failed test "both filters":
message #0: error evaluating matching filters: conflicting filters detected: 'markImportant' is applied differently: got true and false
Note:
- Message: {
    "to": [
      "myalias@gmail.com"
    ],
    "lists": [
      "list2"
    ]
  }
`,
		},
		{
			path:    "fail.jsonnet",
			numErrs: 2,
			expectedOut: `Failed: 2/2

Failed test "wrong test":
message #0: error evaluating matching filters: conflicting filters detected: 'markImportant' is applied differently: got true and false
Note:
- Message: {
    "to": [
      "myalias@gmail.com"
    ],
    "lists": [
      "list2"
    ]
  }

Failed test "another wrong test":
multiple errors (2):
- message #0 is going to get unexpected actions: {"markImportant":false}
  Note:
  - Message: {
      "lists": [
        "list1"
      ]
    }
  - Actions:
    --- want
    +++ got
    @@ -1,3 +1,3 @@
     {
    -  "markImportant": true
    +  "markImportant": false
     }
- message #1 is going to get unexpected actions: {"markImportant":false}
  Note:
  - Message: {
      "lists": [
        "list2"
      ]
    }
  - Actions:
    --- want
    +++ got
    @@ -1,3 +1,3 @@
     {
    -  "markImportant": true
    +  "markImportant": false
     }
`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			cfg := readConfig(t, tc.path)
			pres, err := apply.FromConfig(cfg)
			assert.Nil(t, err)
			rules, errs := NewFromParserRules(pres.Rules)
			assert.Empty(t, errs)
			res := rules.ExecTests(cfg.Tests)
			assert.Equal(t, len(res.Failed), tc.numErrs)
			if tc.expectedOut != "" {
				assert.Equal(t, tc.expectedOut, res.String())
			}
		})
	}
}
