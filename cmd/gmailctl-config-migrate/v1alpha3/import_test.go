package v1alpha3

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	"github.com/mbrt/gmailctl/cmd/gmailctl-config-migrate/v1alpha2"
	"github.com/mbrt/gmailctl/internal/config/v1alpha3"
)

func read(path string) []byte {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return b
}

func parseV2(t *testing.T, path string) v1alpha2.Config {
	var res v1alpha2.Config
	if err := yaml.UnmarshalStrict(read(path), &res); err != nil {
		t.Fatal(err)
	}
	return res
}

func parseV3(t *testing.T, path string) v1alpha3.Config {
	var res v1alpha3.Config
	if err := json.Unmarshal(read(path), &res); err != nil {
		t.Fatal(err)
	}
	return res
}

func dump(t *testing.T, cfg v1alpha3.Config) string {
	b, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func testFiles(t *testing.T, ext string) []string {
	fs, err := filepath.Glob(fmt.Sprintf("testdata/*.%s", ext))
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(fs)
	return fs
}

func TestConvert(t *testing.T) {
	v2files := testFiles(t, "v2.yaml")
	v3files := testFiles(t, "v3.json")
	assert.Len(t, v2files, len(v3files))

	for i := 0; i < len(v2files); i++ {
		t.Run(v2files[i], func(t *testing.T) {
			v2cfg := parseV2(t, v2files[i])
			v3cfg := parseV3(t, v3files[i])

			got, err := Import(v2cfg)
			assert.Nil(t, err)
			assert.Equal(t, dump(t, v3cfg), dump(t, got))
		})
	}
}
