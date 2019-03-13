package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/mbrt/gmailctl/pkg/filter"
)

var (
	diffFilename string
)

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Shows a diff between the local configuaration and Gmail settings",
	Long: `The diff command shows the difference between the local
configuration and the current Gmail settings of your account.

By default diff uses the configuration file inside the config
directory [config.(yaml|jsonnet)].`,
	Run: func(cmd *cobra.Command, args []string) {
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
}

func diff(path string) error {
	parseRes, err := parseConfig(path, "")
	if err != nil {
		return err
	}

	gmailapi, err := openAPI()
	if err != nil {
		return configurationError(errors.Wrap(err, "cannot connect to Gmail"))
	}

	upstream, err := upstreamFilters(gmailapi)
	if err != nil {
		return err
	}

	diff, err := filter.Diff(upstream, parseRes.filters)
	if err != nil {
		return errors.New("cannot compare upstream with local filters")
	}

	fmt.Print(diff)
	return nil
}
