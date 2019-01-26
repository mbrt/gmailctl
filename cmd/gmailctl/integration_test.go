package main

import (
	"flag"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	cfg "github.com/mbrt/gmailctl/pkg/config/v1alpha2"
	"github.com/mbrt/gmailctl/pkg/filter"
	"github.com/mbrt/gmailctl/pkg/parser"
)

// update is useful to regenerate the diff files, whenever necessary.
// Make sure the new version makes sense!!
var update = flag.Bool("update", false, "update .diff files")

func read(t *testing.T, path string) []byte {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func readConfig(t *testing.T, path string) cfg.Config {
	b := read(t, path)
	var res cfg.Config
	if err := yaml.UnmarshalStrict(b, &res); err != nil {
		t.Fatal(err)
	}
	return res
}

type testPaths struct {
	locals []string
	diffs  []string
}

func globPaths(t *testing.T, pattern string) []string {
	fs, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(fs)
	return fs
}

func allTestPaths(t *testing.T) testPaths {
	tp := testPaths{
		locals: globPaths(t, "testdata/local.*.yaml"),
		diffs:  globPaths(t, "testdata/local.*.diff"),
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

func TestIntegration(t *testing.T) {
	remoteFilt, err := cfgPathToFilters(t, "testdata/remote.yaml")
	assert.Nil(t, err)
	tps := allTestPaths(t)

	for i := 0; i < len(tps.locals); i++ {
		local := tps.locals[i]

		t.Run(local, func(t *testing.T) {
			diffFile := tps.diffs[i]
			locFilt, err := cfgPathToFilters(t, local)
			assert.Nil(t, err)

			// Remote filters can come in _any_ order
			// We can make the test more realistic by shuffling them here
			rand.Shuffle(len(remoteFilt), func(i, j int) {
				remoteFilt[i], remoteFilt[j] = remoteFilt[j], remoteFilt[i]
			})

			diff, err := filter.Diff(remoteFilt, locFilt)
			assert.Nil(t, err)

			if *update {
				// Update the golden files
				err = ioutil.WriteFile(diffFile, []byte(diff.String()), 0644)
				assert.Nil(t, err)
			} else {
				// Test them
				expectedDiff := read(t, diffFile)
				assert.Equal(t, string(expectedDiff), diff.String())
			}
		})
	}
}
