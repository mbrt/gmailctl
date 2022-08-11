package v1alpha2

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/mbrt/gmailctl/cmd/gmailctl-config-migrate/v1alpha1"
)

func read(path string) io.Reader {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return bytes.NewBuffer(b)
}

func parseV1(t *testing.T, path string) v1alpha1.Config {
	var res v1alpha1.Config
	dec := yaml.NewDecoder(read(path))
	dec.KnownFields(true)
	if err := dec.Decode(&res); err != nil {
		t.Fatal(err)
	}
	return res
}

func parseV2(t *testing.T, path string) Config {
	var res Config
	dec := yaml.NewDecoder(read(path))
	dec.KnownFields(true)
	if err := dec.Decode(&res); err != nil {
		t.Fatal(err)
	}
	return res
}

func dump(t *testing.T, cfg Config) string {
	b, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func testFiles(t *testing.T, version string) []string {
	fs, err := filepath.Glob(fmt.Sprintf("testdata/*.%s.yaml", version))
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(fs)
	return fs
}

func TestConvert(t *testing.T) {
	v1files := testFiles(t, "v1")
	v2files := testFiles(t, "v2")
	assert.Len(t, v1files, len(v2files))

	for i := 0; i < len(v1files); i++ {
		t.Run(v1files[i], func(t *testing.T) {
			v1cfg := parseV1(t, v1files[i])
			v2cfg := parseV2(t, v2files[i])

			got, err := Import(v1cfg)
			assert.Nil(t, err)
			assert.Equal(t, dump(t, v2cfg), dump(t, got))
		})
	}
}

func TestWrongConsts(t *testing.T) {
	v1c := v1alpha1.Config{
		Consts: v1alpha1.Consts{
			"c1": v1alpha1.ConstValue{Values: []string{"foo"}},
		},
		Rules: []v1alpha1.Rule{
			{
				Filters: v1alpha1.Filters{
					Consts: v1alpha1.CompositeFilters{
						MatchFilters: v1alpha1.MatchFilters{
							From: []string{"c2"},
						},
					},
				},
				Actions: v1alpha1.Actions{},
			},
		},
	}
	_, err := Import(v1c)
	assert.NotNil(t, err)
}
