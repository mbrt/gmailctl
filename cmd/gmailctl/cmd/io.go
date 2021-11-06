package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mbrt/gmailctl/internal/errors"
)

func askYN(prompt string) bool {
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s [y/N]: ", prompt)
		if choice, err := r.ReadString('\n'); err == nil {
			switch strings.ToLower(strings.TrimRight(choice, "\r\n")) {
			case "y", "yes":
				return true
			case "n", "no", "": // empty string defaults to 'no'
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
	stderrPrintf("Error: %v\n", err)
	if det := errors.Details(err); det != "" {
		stderrPrintf("\nNote: %s\n", det)
	}
	os.Exit(1)
}

func stderrPrintf(format string, a ...interface{}) {
	/* #nosec */
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}
