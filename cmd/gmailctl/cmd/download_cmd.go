package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/mbrt/gmailctl/pkg/rimport"
)

var (
	downloadOutput string
)

const downloadHeader = `// Auto-imported filters by 'gmailctl download'.
//
// WARNING: This functionality is experimental. Before making any
// changes, check that no diff is detected with the remote filters by
// using the 'diff' command.

// Uncomment if you want to use the standard library.
// local lib = import 'gmailctl.libsonnet';
`

const labelsComment = `  // Note: labels management is optional. If you prefer to use the
  // GMail interface to add and remove labels, you can safely remove
  // this section of the config.
`

var labelsLine = "  labels: ["

// downloadCmd represents the import command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download filters from Gmail to a local config file",
	Long: `The download command gets the currently configured filters from Gmail
and generates a compatible configuration file.

The resulting config won't be pretty, but it should be a good starting
point if your filters have been managed by other means and you want to
move to gmailctl.

WARNING: This functionality is experimental. After downloading, verify
that no diff is detected with the remote filters by using the 'diff'
command.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := download(downloadOutput); err != nil {
			fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	// Flags and configuration settings
	downloadCmd.PersistentFlags().StringVarP(&downloadOutput, "output", "o", "", "output file (default to stdout)")
}

func download(outputPath string) (err error) {
	var out io.Writer
	if outputPath == "" {
		out = os.Stdout
	} else {
		f, e := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if e != nil {
			return fmt.Errorf("opening output: %w", err)
		}
		defer func() {
			e = f.Close()
			// do not hide more important error
			if err == nil {
				err = e
			}
		}()
		out = f
	}
	return downloadWithOut(out)
}

func downloadWithOut(out io.Writer) error {
	gmailapi, err := openAPI()
	if err != nil {
		return configurationError(fmt.Errorf("connecting to Gmail: %w", err))
	}

	upstream, err := upstreamConfig(gmailapi)
	if err != nil {
		return err
	}

	cfg, err := rimport.Import(upstream.Filters, upstream.Labels)
	if err != nil {
		return err
	}

	err = marshalJsonnet(cfg, out)
	if err != nil {
		return fmt.Errorf("converting to Jsonnet: %w", err)
	}
	return nil
}

func marshalJsonnet(v interface{}, w io.Writer) error {
	// Convert to JSON
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	// Make JSON resemble Jsonnet by getting rid of unnecessary quotes
	reader := bufio.NewReader(bytes.NewReader(b))
	writer := bufio.NewWriter(w)
	keyRe := regexp.MustCompile(`^ *"([a-zA-Z01]+)":`)
	var line []byte

	_, err = writer.WriteString(downloadHeader)
	if err != nil {
		return err
	}

	line, _, err = reader.ReadLine()
	for err == nil {
		line = replaceGroupsRe(keyRe, line)
		if string(line) == labelsLine {
			_, err = writer.WriteString(labelsComment)
			if err != nil {
				break
			}
		}
		_, err = writer.Write(line)
		if err != nil {
			break
		}
		_, err = writer.WriteRune('\n')
		if err != nil {
			break
		}

		line, _, err = reader.ReadLine()
	}

	if err == io.EOF {
		return writer.Flush()
	}
	return err
}

func replaceGroupsRe(re *regexp.Regexp, in []byte) []byte {
	m := re.FindSubmatchIndex(in)
	if len(m) == 0 {
		return in
	}
	keyb, keye := m[2], m[3]

	var res []byte
	res = append(res, in[:keyb-1]...)
	res = append(res, in[keyb:keye]...)
	res = append(res, in[keye+1:]...)
	return res
}
