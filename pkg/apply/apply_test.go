package apply

import (
	"flag"
	"fmt"
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
		t.Fatal("expected both local and diff files to be present")
	}

	// Get the files in the same order
	sort.Strings(tp.locals)
	sort.Strings(tp.diffs)

	return tp
}

func parseConfig(t *testing.T, path string) ConfigParseRes {
	t.Helper()
	config := readConfig(t, path)
	r, err := FromConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func shuffleConfig(cfg GmailConfig) {
	fs := cfg.Filters
	rand.Shuffle(len(fs), func(i, j int) {
		fs[i], fs[j] = fs[j], fs[i]
	})
	ls := cfg.Labels
	rand.Shuffle(len(ls), func(i, j int) {
		ls[i], ls[j] = ls[j], ls[i]
	})
}

func addIDs(cfg GmailConfig) {
	for i := range cfg.Filters {
		cfg.Filters[i].ID = fmt.Sprintf("filter%d", i)
	}
	for i := range cfg.Labels {
		cfg.Labels[i].ID = fmt.Sprintf("label%d", i)
	}
}

func labelIDs(ls label.Labels) []string {
	var ret []string
	for _, l := range ls {
		ret = append(ret, l.ID)
	}
	return ret
}

func filterIDs(fs filter.Filters) []string {
	var ret []string
	for _, f := range fs {
		ret = append(ret, f.ID)
	}
	return ret
}

func toLabels(ms []label.ModifiedLabel) label.Labels {
	var lbs label.Labels
	for _, m := range ms {
		label := m.New
		label.ID = m.Old.ID
		lbs = append(lbs, label)
	}
	return lbs
}

type fakeAPI struct {
	addLabels     label.Labels
	updateLabels  label.Labels
	deleteLabels  []string
	addFilters    filter.Filters
	deleteFilters []string
}

func (f *fakeAPI) AddLabels(lbs label.Labels) error {
	f.addLabels = lbs
	return nil
}

func (f *fakeAPI) AddFilters(fs filter.Filters) error {
	f.addFilters = fs
	return nil
}

func (f *fakeAPI) UpdateLabels(lbs label.Labels) error {
	f.updateLabels = lbs
	return nil
}
func (f *fakeAPI) DeleteFilters(ids []string) error {
	f.deleteFilters = ids
	return nil
}
func (f *fakeAPI) DeleteLabels(ids []string) error {
	f.deleteLabels = ids
	return nil
}

func TestDiff(t *testing.T) {
	upstream := parseConfig(t, filepath.Join(testdataDir, "remote.jsonnet")).GmailConfig
	tps := allTestPaths(t)

	for i := 0; i < len(tps.locals); i++ {
		local := tps.locals[i]

		t.Run(local, func(t *testing.T) {
			diffFile := tps.diffs[i]
			config := parseConfig(t, local)

			// Remote filters and labels can come in _any_ order
			// We can make the test more realistic by shuffling them here
			shuffleConfig(upstream)

			diff, err := Diff(config.GmailConfig, upstream)
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

func TestApply(t *testing.T) {
	upstream := parseConfig(t, filepath.Join(testdataDir, "remote.jsonnet")).GmailConfig
	tps := allTestPaths(t)

	for i := 0; i < len(tps.locals); i++ {
		local := tps.locals[i]

		t.Run(local, func(t *testing.T) {
			config := parseConfig(t, local)
			api := fakeAPI{}

			// Remote filters and labels can come in _any_ order
			// We can make the test more realistic by shuffling them here
			shuffleConfig(upstream)
			addIDs(upstream)

			diff, err := Diff(config.GmailConfig, upstream)
			assert.Nil(t, err)

			// Apply without removing labels
			err = Apply(diff, &api, false)
			assert.Nil(t, err)
			assert.Equal(t, diff.FiltersDiff.Added, api.addFilters)
			assert.Equal(t, filterIDs(diff.FiltersDiff.Removed), api.deleteFilters)
			assert.Equal(t, diff.LabelsDiff.Added, api.addLabels)
			assert.Equal(t, toLabels(diff.LabelsDiff.Modified), api.updateLabels)
			// Labels are never removed
			assert.Empty(t, api.deleteLabels)

			// Apply with removing labels
			if len(diff.LabelsDiff.Removed) > 0 {
				err = Apply(diff, &api, true)
				assert.Nil(t, err)
				assert.Equal(t, labelIDs(diff.LabelsDiff.Removed), api.deleteLabels)
			}

		})
	}
}
