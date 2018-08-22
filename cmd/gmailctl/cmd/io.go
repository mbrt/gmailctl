package cmd

import (
	"fmt"
	"os"
	"strings"
)

func askYN(prompt string) bool {
	for {
		fmt.Printf("%s [y/N]: ", prompt)
		var choice string
		if _, err := fmt.Scan(&choice); err == nil {
			switch strings.ToLower(choice) {
			case "y":
				return true
			case "yes":
				return true
			case "n":
				return false
			case "no":
				return false
			}
		}
		fmt.Println("invalid choice")
	}
}

func fatal(err error) {
	stderrPrintf("%v\n", err)
	if HasUserHelp(err) {
		stderrPrintf("\nNote: %s\n", GetUserHelp(err))
	}
	os.Exit(1)
}

func stderrPrintf(format string, a ...interface{}) {
	/* #nosec */
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}
