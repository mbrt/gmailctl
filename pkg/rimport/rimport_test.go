package rimport

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	"github.com/mbrt/gmailctl/pkg/filter"
)

// update is useful to regenerate the yaml files, whenever necessary.
// Make sure the new version makes sense!!
var update = flag.Bool("update", false, "update .yaml files")

func read(t *testing.T, path string) []byte {
	t.Helper()
	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func readFilters(t *testing.T, path string) filter.Filters {
	t.Helper()
	b := read(t, path)
	var res filter.Filters
	err := json.Unmarshal(b, &res)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestImport(t *testing.T) {
	remote := readFilters(t, "testdata/remote.json")

	imp, err := Import(remote)
	assert.Nil(t, err)
	b, err := yaml.Marshal(imp)
	assert.Nil(t, err)

	if *update {
		// Update the golden files
		err = ioutil.WriteFile("testdata/config.yaml", []byte(b), 0644)
		assert.Nil(t, err)
	} else {
		// Test them
		expected := read(t, "testdata/config.yaml")
		assert.Equal(t, string(expected), string(b))
	}
}
