package cfgtest

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mbrt/gmailctl/pkg/apply"
	cfg "github.com/mbrt/gmailctl/pkg/config"
	cfgv3 "github.com/mbrt/gmailctl/pkg/config/v1alpha3"
)

func read(t *testing.T, path string) []byte {
	t.Helper()
	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func readConfig(t *testing.T, path string) cfgv3.Config {
	t.Helper()
	path = filepath.Join("testdata", path)
	res, err := cfg.ReadFile(path, path)
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
