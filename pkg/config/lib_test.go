package config

import (
	"flag"
	"io/ioutil"
	"path/filepath"
	"sort"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

// update is useful to regenerate the diff files, whenever necessary.
// Make sure the new version makes sense!!
var update = flag.Bool("update", false, "update .diff files")

type testPaths struct {
	jsonnets []string
	yamls    []string
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
		jsonnets: globPaths(t, "testdata/*.jsonnet"),
		yamls:    globPaths(t, "testdata/*.yaml"),
	}
	if len(tp.jsonnets) != len(tp.yamls) {
		t.Fatal("expected both jsonnet and yaml to be present")
	}
	return tp
}

func TestJsonnetLib(t *testing.T) {
	tps := allTestPaths(t)
	for i := 0; i < len(tps.jsonnets); i++ {
		jfile := tps.jsonnets[i]

		t.Run(jfile, func(t *testing.T) {
			jparsed, err := ReadFile(jfile)
			assert.Nil(t, err)

			yfile := tps.yamls[i]
			if *update {
				// Update the golden files
				buf, err := yaml.Marshal(jparsed)
				assert.Nil(t, err)
				err = ioutil.WriteFile(yfile, buf, 0644)
				assert.Nil(t, err)
			} else {
				// Test them
				yparsed, err := ReadFile(yfile)
				assert.Nil(t, err)
				assert.Equal(t, yparsed, jparsed)
			}
		})
	}
}
