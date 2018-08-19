package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/mbrt/gmailfilter/pkg/config"
	export "github.com/mbrt/gmailfilter/pkg/export/xml"
	"github.com/mbrt/gmailfilter/pkg/filter"
)

func readConfig(path string) (config.Config, error) {
	/* #nosec */
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return config.Config{}, errors.Wrap(err, fmt.Sprintf("cannot read %s", path))
	}

	var res config.Config
	err = yaml.Unmarshal(b, &res)
	return res, err

}

func errorf(format string, a ...interface{}) {
	/* #nosec */
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}

func fatal(format string, a ...interface{}) {
	errorf(format, a...)
	if !strings.HasSuffix(format, "\n") {
		errorf("\n") // Add newline
	}
	os.Exit(1)
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fatal("usage: %s <config-file>", os.Args[0])
	}
	cfg, err := readConfig(flag.Arg(0))
	if err != nil {
		fatal("error in config parse: %s", err)
	}

	filters, err := filter.FromConfig(cfg)
	if err != nil {
		fatal("errors exporting filters: %s", err)
	}

	err = export.DefaultExporter().Export(cfg.Author, filters, os.Stdout)
	if err != nil {
		fatal("error exporting to XML: %s", err)
	}
}
