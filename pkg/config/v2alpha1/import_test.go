package v2alpha1

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	v1 "github.com/mbrt/gmailctl/pkg/config/v1alpha1"
)

func read(path string) []byte {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return b
}

func parseV1(t *testing.T, path string) v1.Config {
	var res v1.Config
	if err := yaml.UnmarshalStrict(read(path), &res); err != nil {
		t.Fatal(err)
	}
	return res
}

func parseV2(t *testing.T, path string) Config {
	var res Config
	if err := yaml.UnmarshalStrict(read(path), &res); err != nil {
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

	if l1, l2 := len(v1files), len(v2files); l1 != l2 {
		t.Fatalf("len(v1files) = %d, want %d", l1, l2)
	}

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
	v1c := v1.Config{
		Consts: v1.Consts{
			"c1": v1.ConstValue{Values: []string{"foo"}},
		},
		Rules: []v1.Rule{
			{
				Filters: v1.Filters{
					Consts: v1.CompositeFilters{
						MatchFilters: v1.MatchFilters{
							From: []string{"c2"},
						},
					},
				},
				Actions: v1.Actions{},
			},
		},
	}
	_, err := Import(v1c)
	assert.NotNil(t, err)
}
