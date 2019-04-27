package v1alpha3

// import (
// 	"fmt"
// 	"io/ioutil"
// 	"path/filepath"
// 	"sort"
// 	"testing"
//
// 	"github.com/stretchr/testify/assert"
// 	"gopkg.in/yaml.v2"
//
// 	v2 "github.com/mbrt/gmailctl/pkg/config/v1alpha2"
// )

// func read(path string) []byte {
// 	b, err := ioutil.ReadFile(path)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return b
// }
//
// func parseV1(t *testing.T, path string) v2.Config {
// 	var res v2.Config
// 	if err := yaml.UnmarshalStrict(read(path), &res); err != nil {
// 		t.Fatal(err)
// 	}
// 	return res
// }
//
// func parseV2(t *testing.T, path string) Config {
// 	var res Config
// 	if err := yaml.UnmarshalStrict(read(path), &res); err != nil {
// 		t.Fatal(err)
// 	}
// 	return res
// }
//
// func dump(t *testing.T, cfg Config) string {
// 	b, err := yaml.Marshal(cfg)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	return string(b)
// }
//
// func testFiles(t *testing.T, version string) []string {
// 	fs, err := filepath.Glob(fmt.Sprintf("testdata/*.%s.yaml", version))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	sort.Strings(fs)
// 	return fs
// }
//
// func TestConvert(t *testing.T) {
// 	v1files := testFiles(t, "v2")
// 	v2files := testFiles(t, "v2")
// 	assert.Len(t, v1files, len(v2files))
//
// 	for i := 0; i < len(v1files); i++ {
// 		t.Run(v1files[i], func(t *testing.T) {
// 			v1cfg := parseV1(t, v1files[i])
// 			v2cfg := parseV2(t, v2files[i])
//
// 			got, err := Import(v1cfg)
// 			assert.Nil(t, err)
// 			assert.Equal(t, dump(t, v2cfg), dump(t, got))
// 		})
// 	}
// }
//
// func TestWrongConsts(t *testing.T) {
// 	v1c := v2.Config{
// 		Consts: v2.Consts{
// 			"c1": v2.ConstValue{Values: []string{"foo"}},
// 		},
// 		Rules: []v2.Rule{
// 			{
// 				Filters: v2.Filters{
// 					Consts: v2.CompositeFilters{
// 						MatchFilters: v2.MatchFilters{
// 							From: []string{"c2"},
// 						},
// 					},
// 				},
// 				Actions: v2.Actions{},
// 			},
// 		},
// 	}
// 	_, err := Import(v1c)
// 	assert.NotNil(t, err)
// }
