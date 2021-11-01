// gmailctl-config-migrate is a tool to migrate gmailctl YAML configs to the latest version.
package main

import (
	"flag"
	"fmt"
	"os"
)

func do(path string) error {
	return importConfig(path, os.Stdout)
}

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Fprintln(os.Stderr, "No config file specified")
		os.Exit(1)
	}
	for _, path := range flag.Args() {
		if err := do(path); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}
