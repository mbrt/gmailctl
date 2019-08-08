package main

import (
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mbrt/gmailctl/pkg/apply"
	cfg "github.com/mbrt/gmailctl/pkg/config"
	cfgv3 "github.com/mbrt/gmailctl/pkg/config/v1alpha3"
	"github.com/mbrt/gmailctl/pkg/filter"
	"github.com/mbrt/gmailctl/pkg/parser"
	"github.com/mbrt/gmailctl/pkg/rimport"
)

const testdataDir = "../../testdata"

func readConfig(t *testing.T, path string) cfgv3.Config {
	t.Helper()
	res, err := cfg.ReadFile(path, filepath.Join(testdataDir, path))
	if err != nil {
		t.Fatal(err)
	}
	return res
}

type testPaths struct {
	locals []string
	diffs  []string
}

func globTestdataPaths(t *testing.T, pattern string) []string {
	t.Helper()
	fs, err := filepath.Glob(filepath.Join(testdataDir, pattern))
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(fs)
	return fs
}

func allTestPaths(t *testing.T) testPaths {
	t.Helper()
	local := globTestdataPaths(t, "local.*.yaml")
	local = append(local, globTestdataPaths(t, "local.*.jsonnet")...)
	tp := testPaths{
		locals: local,
		diffs:  globTestdataPaths(t, "local.*.diff"),
	}
	if len(tp.locals) != len(tp.diffs) {
		t.Fatal("expected both yaml and diff to be present")
	}
	return tp
}

func cfgPathToFilters(t *testing.T, path string) (filter.Filters, error) {
	t.Helper()
	config := readConfig(t, path)
	rules, err := parser.Parse(config)
	if err != nil {
		return filter.Filters{}, err
	}
	return filter.FromRules(rules)
}

func TestIntegrationImport(t *testing.T) {
	tps := allTestPaths(t)

	for i := 0; i < len(tps.locals); i++ {
		local := tps.locals[i]

		t.Run(local, func(t *testing.T) {
			locFilt, err := cfgPathToFilters(t, local)
			assert.Nil(t, err)

			// Import
			config, err := rimport.Import(locFilt)
			assert.Nil(t, err)
			// Generate
			diff, err := apply.Diff(config, apply.GmailConfig{Filters: locFilt})
			// Re-generating imported filters should not cause any diff
			assert.Nil(t, err)
			assert.Equal(t, "", diff.String())
		})
	}
}
