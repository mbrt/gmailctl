package label

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func loadLabels(t *testing.T, path string) Labels {
	t.Helper()
	res := Labels{}
	err := json.Unmarshal(read(t, path), &res)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

type testPaths struct {
	locals []string
	diffs  []string
}

func globPaths(t *testing.T, pattern string) []string {
	t.Helper()
	ls, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(ls)
	return ls
}

func allTestPaths(t *testing.T) testPaths {
	t.Helper()
	local := globPaths(t, "testdata/local.*.json")
	tp := testPaths{
		locals: local,
		diffs:  globPaths(t, "testdata/local.*.diff"),
	}
	if len(tp.locals) != len(tp.diffs) {
		t.Fatal("expected both json and diff files to be present")
	}
	return tp
}

func TestDiff(t *testing.T) {
	upstream := loadLabels(t, "testdata/upstream.json")
	tps := allTestPaths(t)

	for i := 0; i < len(tps.locals); i++ {
		localPath := tps.locals[i]

		t.Run(localPath, func(t *testing.T) {
			diffPath := tps.diffs[i]
			local := loadLabels(t, localPath)

			// Remote labels can come in _any_ order
			// We can make the test more realistic by shuffling them here
			rand.Shuffle(len(upstream), func(i, j int) {
				upstream[i], upstream[j] = upstream[j], upstream[i]
			})
			diff, err := Diff(upstream, local)
			assert.Nil(t, err)

			if *update {
				// Update the golden files
				err = ioutil.WriteFile(diffPath, []byte(diff.String()), 0644)
				assert.Nil(t, err)
			} else {
				// Test them
				expectedDiff := read(t, diffPath)
				assert.Equal(t, string(expectedDiff), diff.String())
			}
		})
	}
}
