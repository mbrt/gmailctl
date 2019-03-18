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
		if _, err := fmt.Scanln(&choice); err == nil {
			switch strings.ToLower(choice) {
			case "y", "yes":
				return true
			case "n", "no":
				return false
			default:
				return false
			}
		}
		fmt.Println("invalid choice")
	}
}

func askOptions(prompt string, choices []string) int {
	var prettyChoices []string
	for _, c := range choices {
		if len(c) == 0 {
			panic("unexpected empty choice")
		}
		p := fmt.Sprintf("[%s] %s", string(c[0]), c)
		prettyChoices = append(prettyChoices, p)
	}

	for {
		fmt.Printf("%s:\n", prompt)
		for _, c := range prettyChoices {
			fmt.Printf("    %s\n", c)
		}
		fmt.Printf("> ")

		var choice string
		if _, err := fmt.Scanln(&choice); err == nil {
			choice = strings.ToLower(choice)
			for i, c := range choices {
				if strings.HasPrefix(c, choice) {
					return i
				}
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
