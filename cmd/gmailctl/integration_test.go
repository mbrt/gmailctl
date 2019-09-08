package main

import (
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mbrt/gmailctl/pkg/apply"
	cfg "github.com/mbrt/gmailctl/pkg/config"
	cfgv3 "github.com/mbrt/gmailctl/pkg/config/v1alpha3"
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

func parseConfig(t *testing.T, path string) apply.ConfigParseRes {
	t.Helper()
	config := readConfig(t, path)
	r, err := apply.FromConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestIntegrationImport(t *testing.T) {
	tps := allTestPaths(t)

	for i := 0; i < len(tps.locals); i++ {
		localPath := tps.locals[i]

		t.Run(localPath, func(t *testing.T) {
			local := parseConfig(t, localPath).GmailConfig

			// Import
			config, err := rimport.Import(local.Filters, local.Labels)
			assert.Nil(t, err)
			// Generate
			pres, err := apply.FromConfig(config)
			assert.Nil(t, err)
			diff, err := apply.Diff(pres.GmailConfig, local)
			// Re-generating imported filters should not cause any diff
			assert.Nil(t, err)
			assert.Equal(t, "", diff.String())
		})
	}
}
