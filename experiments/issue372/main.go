package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/go-jsonnet"

	"github.com/mbrt/gmailctl/internal/data"
	cfg "github.com/mbrt/gmailctl/internal/engine/config/v1alpha3"
	"github.com/mbrt/gmailctl/internal/engine/export/xml"
	"github.com/mbrt/gmailctl/internal/engine/filter"
)

type settings struct {
	Compact bool `json:"compact,omitempty"`
}

// The experiment extends only its own config reader. The production binary
// still rejects settings.compact until a production design is implemented.
type experimentConfig struct {
	cfg.Config
	Settings settings `json:"settings,omitempty"`
}

type experimentImporter struct{ files jsonnet.FileImporter }

func (i *experimentImporter) Import(from, imported string) (jsonnet.Contents, string, error) {
	if imported == "gmailctl.libsonnet" {
		return jsonnet.MakeContents(data.GmailctlLib()), "embedded/gmailctl.libsonnet", nil
	}
	return i.files.Import(from, imported)
}

func readConfig(path string) (experimentConfig, error) {
	var c experimentConfig
	vm := jsonnet.MakeVM()
	vm.Importer(&experimentImporter{files: jsonnet.FileImporter{JPaths: []string{filepath.Dir(path)}}})
	contents, err := vm.EvaluateFile(path)
	if err != nil {
		return c, err
	}
	dec := json.NewDecoder(bytes.NewBufferString(contents))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&c); err != nil {
		return c, err
	}
	if c.Version != cfg.Version {
		return c, fmt.Errorf("expected version %s, got %q", cfg.Version, c.Version)
	}
	return c, nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", "", "Jsonnet config (omit to run the comparison scenarios)")
	strategy := flag.String("strategy", "hysteresis", "hysteresis or toggle")
	operation := flag.String("operation", "plan", "plan, diff, export, or snapshot")
	upstreamPath := flag.String("upstream", "", "JSON snapshot of installed filters (default: empty account)")
	fixtures := flag.String("fixtures", "testdata/valid", "fixture directory for the comparison report")
	flag.Parse()
	if *configPath == "" {
		return compare(os.Stdout, *fixtures)
	}
	if *operation != "plan" && *operation != "diff" && *operation != "export" && *operation != "snapshot" {
		return fmt.Errorf("unknown operation %q", *operation)
	}
	config, err := readConfig(*configPath)
	if err != nil {
		return err
	}
	c, err := generate(config.Config)
	if err != nil {
		return err
	}
	var upstream filter.Filters
	// Export intentionally disregards the snapshot and hysteresis.
	if *upstreamPath != "" && *operation != "export" {
		contents, err := os.ReadFile(*upstreamPath)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(contents, &upstream); err != nil {
			return err
		}
	}
	var d decision
	switch *strategy {
	case "hysteresis":
		d = chooseHysteresis(c, upstream, *operation == "export")
	case "toggle":
		d = chooseToggle(config.Settings.Compact)
	default:
		return fmt.Errorf("unknown strategy %q", *strategy)
	}
	planNotice(os.Stderr, c, d, upstream, *operation == "export")
	desired := d.Filters(c)
	switch *operation {
	case "export":
		return xml.DefaultExporter().Export(config.Author, desired, os.Stdout)
	case "snapshot":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(desired)
	case "diff":
		diff, err := filter.Diff(upstream, desired, false, 3, false)
		if err != nil {
			return err
		}
		fmt.Print(diff)
	default:
		delta := changeCount(upstream, desired)
		fmt.Printf("%s: %d filters; %s; peak during current apply order: %d\n", d.Mode(), len(desired), delta, len(upstream)+delta.Added)
	}
	return nil
}
