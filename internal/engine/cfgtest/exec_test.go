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
		path string
		pass bool
	}{
		{path: "pass.jsonnet", pass: true},
		{path: "invalid.jsonnet", pass: false},
		{path: "fail.jsonnet", pass: false},
	}
	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			cfg := readConfig(t, tc.path)
			pres, err := apply.FromConfig(cfg)
			assert.Nil(t, err)
			rules, errs := NewFromParserRules(pres.Rules)
			assert.Empty(t, errs)
			err = rules.ExecTests(cfg.Tests)
			if tc.pass {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
