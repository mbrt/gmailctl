package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mbrt/gmailctl/internal/errors"
)

func askYN(in io.Reader, out io.Writer, prompt string) bool {
	r := bufio.NewReader(in)
	for {
		fmt.Fprintf(out, "%s [y/N]: ", prompt)
		choice, err := r.ReadString('\n')
		answer := strings.ToLower(strings.TrimRight(choice, "\r\n"))
		if err != nil {
			// Stdin is exhausted (e.g. EOF), so asking again would loop forever.
			// Accept a last unterminated answer, otherwise default to 'no'.
			fmt.Fprintln(out)
			return answer == "y" || answer == "yes"
		}
		switch answer {
		case "y", "yes":
			return true
		case "n", "no", "": // empty string defaults to 'no'
			return false
		}
		fmt.Fprintln(out, "invalid choice")
	}
}

// askOptions returns the index of the choice picked by the user, or an error
// if the input ends before a valid choice is made.
func askOptions(in io.Reader, out io.Writer, prompt string, choices []string) (int, error) {
	var prettyChoices []string
	for _, c := range choices {
		if len(c) == 0 {
			panic("unexpected empty choice")
		}
		p := fmt.Sprintf("[%s] %s", string(c[0]), c)
		prettyChoices = append(prettyChoices, p)
	}

	for {
		fmt.Fprintf(out, "%s:\n", prompt)
		for _, c := range prettyChoices {
			fmt.Fprintf(out, "    %s\n", c)
		}
		fmt.Fprintf(out, "> ")

		var choice string
		_, err := fmt.Fscanln(in, &choice)
		if errors.Is(err, io.EOF) {
			fmt.Fprintln(out)
			return 0, fmt.Errorf("reading answer: %w", err)
		}
		if err == nil {
			choice = strings.ToLower(choice)
			for i, c := range choices {
				if strings.HasPrefix(c, choice) {
					return i, nil
				}
			}
		}

		fmt.Fprintln(out, "invalid choice")
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
