package config

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	cfgv1 "github.com/mbrt/gmailctl/pkg/config/v1alpha3"
)

// update is useful to regenerate the json files, whenever necessary.
// Make sure the new version makes sense!!
var update = flag.Bool("update", false, "update .json files")

type testPaths struct {
	jsonnets []string
	jsons    []string
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
		jsons:    globPaths(t, "testdata/*.json"),
	}
	if len(tp.jsonnets) != len(tp.jsons) {
		t.Fatal("expected both jsonnet and json to be present")
	}
	return tp
}

func read(t *testing.T, path string) []byte {
	t.Helper()
	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestJsonnetLib(t *testing.T) {
	tps := allTestPaths(t)
	for i := 0; i < len(tps.jsonnets); i++ {
		jfile := tps.jsonnets[i]

		t.Run(jfile, func(t *testing.T) {
			jnparsed, err := ReadFile(jfile, "")
			assert.Nil(t, err)

			jsfile := tps.jsons[i]
			if *update {
				// Update the golden files
				buf, err := json.MarshalIndent(jnparsed, "", "  ")
				assert.Nil(t, err)
				err = ioutil.WriteFile(jsfile, buf, 0644)
				assert.Nil(t, err)
			} else {
				// Test them
				var jsparsed cfgv1.Config
				err := json.Unmarshal(read(t, jsfile), &jsparsed)
				assert.Nil(t, err)
				assert.Equal(t, jsparsed, jnparsed)
			}
		})
	}
}
