package apply

import (
	"flag"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	cfg "github.com/mbrt/gmailctl/pkg/config"
	cfgv3 "github.com/mbrt/gmailctl/pkg/config/v1alpha3"
	"github.com/mbrt/gmailctl/pkg/filter"
	"github.com/mbrt/gmailctl/pkg/label"
	"github.com/mbrt/gmailctl/pkg/parser"
)

const testdataDir = "../../testdata"

// update is useful to regenerate the diff files, whenever necessary.
// Make sure the new version makes sense!!
var update = flag.Bool("update", false, "update .diff files")

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

func cfgPathToGmailConfig(t *testing.T, path string) (GmailConfig, error) {
	t.Helper()
	config := readConfig(t, path)
	rules, err := parser.Parse(config)
	if err != nil {
		return GmailConfig{}, err
	}
	fs, err := filter.FromRules(rules)
	labels := label.FromConfig(config.Labels)
	return GmailConfig{
		Filters: fs,
		Labels:  labels,
	}, err
}

func TestDiff(t *testing.T) {
	upstream, err := cfgPathToGmailConfig(t, filepath.Join(testdataDir, "remote.jsonnet"))
	assert.Nil(t, err)
	tps := allTestPaths(t)

	for i := 0; i < len(tps.locals); i++ {
		local := tps.locals[i]

		t.Run(local, func(t *testing.T) {
			diffFile := tps.diffs[i]
			config := readConfig(t, local)

			// Remote filters can come in _any_ order
			// We can make the test more realistic by shuffling them here
			remoteFilt := upstream.Filters
			rand.Shuffle(len(remoteFilt), func(i, j int) {
				remoteFilt[i], remoteFilt[j] = remoteFilt[j], remoteFilt[i]
			})

			diff, err := Diff(config, upstream)
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
