package rimport

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mbrt/gmailctl/pkg/filter"
)

// update is useful to regenerate the json files, whenever necessary.
// Make sure the new version makes sense!!
var update = flag.Bool("update", false, "update golden .json files")

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

	// TODO: Test labels import
	imp, err := Import(remote, nil)
	assert.Nil(t, err)
	b, err := json.MarshalIndent(imp, "", "  ")
	assert.Nil(t, err)

	if *update {
		// Update the golden files
		err = ioutil.WriteFile("testdata/config.json", b, 0644)
		assert.Nil(t, err)
	} else {
		// Test them
		expected := read(t, "testdata/config.json")
		assert.Equal(t, string(expected), string(b))
	}
}
