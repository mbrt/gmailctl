package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	papply "github.com/mbrt/gmailctl/pkg/apply"
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

By default apply uses the configuration file inside the config
directory [config.(yaml|jsonnet)].`,
	Run: func(cmd *cobra.Command, args []string) {
		f := applyFilename
		if f == "" {
			f = configFilenameFromDir(cfgDir)
		}
		if err := apply(f, !applyYes); err != nil {
			fatal(err)
		}
	}}

func init() {
	rootCmd.AddCommand(applyCmd)

	// Flags and configuration settings
	applyCmd.PersistentFlags().StringVarP(&applyFilename, "filename", "f", "", "configuration file")
	applyCmd.Flags().BoolVarP(&applyYes, "yes", "y", false, "don't ask for confirmation, just apply")
}

func apply(path string, interactive bool) error {
	parseRes, err := parseConfig(path, "")
	if err != nil {
		return err
	}

	gmailapi, err := openAPI()
	if err != nil {
		return configurationError(errors.Wrap(err, "cannot connect to Gmail"))
	}

	upstream, err := upstreamConfig(gmailapi)
	if err != nil {
		if err != errLabelsDisabled {
			return err
		}
		// Drop the labels, to be sure we don't try to apply them later on
		parseRes.labels = nil
	}

	diff, err := papply.Diff(parseRes.config, upstream)
	if err != nil {
		return errors.Wrap(err, "cannot compare upstream with local config")
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
	return papply.Apply(diff, gmailapi)
}

func configurationError(err error) error {
	return UserError(err, "The configuration can be initialized with 'gmailctl init'")
}
