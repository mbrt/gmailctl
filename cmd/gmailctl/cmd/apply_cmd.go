package cmd

import (
	"fmt"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/mbrt/gmailctl/pkg/api"
	"github.com/mbrt/gmailctl/pkg/filter"
)

var (
	applyFilename string
	applyYes      bool
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a configuration file to Gmail settings",
	Long: `The apply command applies minimal changes to your Gmail settings
to make them match your local configuration file.

By default apply uses the "config.yaml" file inside the config directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		f := path.Join(cfgDir, "config.yaml")
		if applyFilename != "" {
			f = applyFilename
		}
		if err := apply(f, !applyYes); err != nil {
			fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)

	// Flags and configuration settings
	applyCmd.PersistentFlags().StringVarP(&applyFilename, "filename", "f", "", "configuration file")
	applyCmd.Flags().BoolVarP(&applyYes, "yes", "y", false, "don't ask for confirmation, just apply")
}

func apply(path string, interactive bool) error {
	parseRes, err := parseConfig(path)
	if err != nil {
		return err
	}

	gmailapi, err := openAPI()
	if err != nil {
		return configurationError(errors.Wrap(err, "cannot connect to Gmail"))
	}

	upstreamFilters, err := gmailapi.ListFilters()
	if err != nil {
		return errors.Wrap(err, "cannot get filters from Gmail")
	}

	diff, err := filter.Diff(upstreamFilters, parseRes.filters)
	if err != nil {
		return errors.New("cannot compare upstream with local filters")
	}

	if diff.Empty() {
		fmt.Println("No changes have been made.")
		return nil
	}

	fmt.Printf("You are going to apply the following changes to your settings:\n\n%s", diff)
	if interactive && !askYN("Do you want to apply them?") {
		return nil
	}

	fmt.Println("Applying the changes...")
	return updateFilters(gmailapi, diff)
}

func updateFilters(gmailapi api.GmailAPI, diff filter.FiltersDiff) error {
	if len(diff.Added) > 0 {
		if err := gmailapi.AddFilters(diff.Added); err != nil {
			return errors.Wrap(err, "error adding filters")
		}
	}
	if len(diff.Removed) == 0 {
		return nil
	}

	removedIds := make([]string, len(diff.Removed))
	for i, f := range diff.Removed {
		removedIds[i] = f.ID
	}
	err := gmailapi.DeleteFilters(removedIds)
	return errors.Wrap(err, "error deleting filters")
}

func configurationError(err error) error {
	return UserError(err, "The configuration can be initialized with 'gmailctl init'")
}
