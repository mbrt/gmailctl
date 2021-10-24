package integration_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mbrt/gmailctl/internal/fakegmail"
	"github.com/mbrt/gmailctl/pkg/api"
	"github.com/mbrt/gmailctl/pkg/apply"
	"github.com/mbrt/gmailctl/pkg/config"
	"github.com/mbrt/gmailctl/pkg/config/v1alpha2"
	"github.com/mbrt/gmailctl/pkg/export/xml"
	"github.com/mbrt/gmailctl/pkg/rimport"
)

// update is useful to regenerate the golden files
// Make sure the new version makes sense!!
var update = flag.Bool("update", false, "update golden files")

var fixedTime = mustParseTime("2006-01-02 15:04", "2018-03-08 17:00")

func globTestdataPaths(t *testing.T, pattern string) []string {
	t.Helper()
	fs, err := filepath.Glob(filepath.Join("testdata", pattern))
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(fs)
	return fs
}

func TestIntegration(t *testing.T) {
	cfgPaths := globTestdataPaths(t, "valid/*.jsonnet")
	gapi := api.NewFromService(fakegmail.NewService(context.Background(), t))

	for _, cfgPath := range cfgPaths {
		name := strings.TrimSuffix(cfgPath, ".jsonnet")
		t.Run(name, func(t *testing.T) {
			// Parse the config.
			cfg, err := config.ReadFile(cfgPath, filepath.Join("testdata", cfgPath))
			require.Nil(t, err)
			pres, err := apply.FromConfig(cfg)
			require.Nil(t, err)

			// Export.
			xmlexp := xml.NewWithTime(func() time.Time { return fixedTime })
			var cfgxml bytes.Buffer
			bw := bufio.NewWriter(&cfgxml)
			author := v1alpha2.Author{Name: "Me", Email: "me@gmail.com"}
			err = xmlexp.Export(author, pres.Filters, bw)
			bw.Flush() // Make sure everything is written out.
			require.Nil(t, err)

			// Fetch the upstream filters.
			upres, err := apply.FromAPI(gapi)
			require.Nil(t, err)

			// Apply the diff.
			d, err := apply.Diff(pres.GmailConfig, upres)
			require.Nil(t, err)
			err = apply.Apply(d, gapi, true)
			require.Nil(t, err)

			// Import.
			upres, err = apply.FromAPI(gapi)
			require.Nil(t, err)
			icfg, err := rimport.Import(upres.Filters, upres.Labels)
			require.Nil(t, err)
			icfgJson, err := json.MarshalIndent(icfg, "", "  ")
			require.Nil(t, err)

			// Compare the results with the golden files (or update the golden files).
			if *update {
				// Import.
				err := ioutil.WriteFile(name+".json", icfgJson, 0o644)
				require.Nil(t, err)
				// Diff.
				err = ioutil.WriteFile(name+".diff", []byte(d.String()), 0o644)
				require.Nil(t, err)
				// Export.
				err = ioutil.WriteFile(name+".xml", cfgxml.Bytes(), 0o644)
				require.Nil(t, err)
				return
			}
			// Import.
			b, err := ioutil.ReadFile(name + ".json")
			require.Nil(t, err)
			assert.Equal(t, string(b), string(icfgJson))
			// Diff.
			b, err = ioutil.ReadFile(name + ".diff")
			require.Nil(t, err)
			assert.Equal(t, string(b), d.String())
			// Export
			b, err = ioutil.ReadFile(name + ".xml")
			require.Nil(t, err)
			assert.Equal(t, string(b), cfgxml.String())
		})
	}
}

func mustParseTime(layout, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return t
}
