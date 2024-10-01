package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/mbrt/gmailctl/internal/engine/rimport"
)

const downloadHeader = `// Auto-imported filters by 'gmailctl download'.
//
// WARNING: This functionality is experimental. Before making any
// changes, check that no diff is detected with the remote filters by
// using the 'diff' command.

// Uncomment if you want to use the standard library.
// local lib = import 'gmailctl.libsonnet';
`

var (
	downloadOutput string
)

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
	Run: func(*cobra.Command, []string) {
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

	err = rimport.MarshalJsonnet(cfg, out, downloadHeader)
	if err != nil {
		return fmt.Errorf("converting to Jsonnet: %w", err)
	}
	return nil
}
