package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	papply "github.com/mbrt/gmailctl/internal/engine/apply"
	"github.com/mbrt/gmailctl/internal/errors"
)

var (
	diffFilename string
	diffDebug    bool
	diffContext  int
)

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Shows a diff between the local configuration and Gmail settings",
	Long: `The diff command shows the difference between the local
configuration and the current Gmail settings of your account.

By default diff uses the configuration file inside the config
directory [config.jsonnet].`,
	Run: func(*cobra.Command, []string) {
		f := diffFilename
		if f == "" {
			f = configFilenameFromDir(cfgDir)
		}
		if err := diff(f); err != nil {
			fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)

	// Flags and configuration settings
	diffCmd.PersistentFlags().StringVarP(&diffFilename, "filename", "f", "", "configuration file")
	diffCmd.PersistentFlags().BoolVar(&diffDebug, "debug", false, "print extra debugging information")
	diffCmd.PersistentFlags().IntVar(&diffContext, "context", papply.DefaultContextLines, "number of lines of filter diff context to show")
}

func diff(path string) error {
	if diffContext < 0 {
		return errors.New("--context must be non-negative")
	}

	useColor := shouldUseColorDiff()

	parseRes, err := parseConfig(path, "", false)
	if err != nil {
		return err
	}

	gmailapi, err := openAPI()
	if err != nil {
		return configurationError(fmt.Errorf("cannot connect to Gmail: %w", err))
	}

	upstream, err := upstreamConfig(gmailapi)
	if err != nil {
		return err
	}

	diff, err := papply.Diff(parseRes.Res.GmailConfig, upstream, diffDebug, diffContext, useColor)
	if err != nil {
		return fmt.Errorf("cannot compare upstream with local config: %w", err)
	}

	fmt.Print(diff)
	return nil
}
