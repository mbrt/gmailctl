package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

func readConfig(path string) (Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var res Config
	err = yaml.Unmarshal(b, &res)
	return res, err

}

func fatal(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	if !strings.HasSuffix(format, "\n") {
		// Add newline
		fmt.Fprintln(os.Stderr, "")
	}
	os.Exit(1)
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fatal("usage: %s <config-file>", os.Args[0])
	}
	cfg, err := readConfig("example.yaml")
	if err != nil {
		fatal("error in config parse: %s", err)
	}

	rules, err := GenerateRules(cfg)
	if err != nil {
		fatal("error generating rules: %s", err)
	}

	err = DefaultXMLExporter().MarshalEntries(cfg.Author, rules, os.Stdout)
	if err != nil {
		fatal("error exporting to XML: %s", err)
	}
}
