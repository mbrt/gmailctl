package cmd

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/mbrt/gmailfilter/pkg/api"
	"github.com/mbrt/gmailfilter/pkg/config"
	"github.com/mbrt/gmailfilter/pkg/filter"
)

var (
	applyFilename string
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
		if err := apply(f); err != nil {
			fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)

	// Flags and configuration settings
	applyCmd.PersistentFlags().StringVarP(&applyFilename, "filename", "f", "", "configuration file")
}

func apply(path string) error {
	cfg, err := config.ParseFile(path)
	if err != nil {
		if config.IsNotFound(err) {
			return configurationError(err)
		}
		return errors.Wrap(err, "cannot parse config file")
	}

	newFilters, err := filter.FromConfig(cfg)
	if err != nil {
		fatal(errors.Wrap(err, "error exporting local filters"))
	}

	gmailapi, err := openAPI()
	if err != nil {
		fatal(errors.Wrap(err, "cannot connect to Gmail"))
	}

	upstreamFilters, err := gmailapi.ListFilters()
	if err != nil {
		fatal(errors.Wrap(err, "cannot get filters from Gmail"))
	}

	diff, err := filter.Diff(upstreamFilters, newFilters)
	if err != nil {
		fatal(errors.New("cannot compare upstream with local filters"))
	}

	if diff.Empty() {
		fmt.Println("No changes have been made.")
		return nil
	}

	fmt.Printf("You are going to apply the following changes to your settings:\n\n%s", diff)
	if !askYN("Do you want to apply them?") {
		return nil
	}

	if err = updateFilters(gmailapi, diff); err != nil {
		fatal(err)
	}

	return nil
}

func openAPI() (api.GmailAPI, error) {
	cred, err := os.Open("credentials.json")
	if err != nil {
		return nil, configurationError(errors.Wrap(err, "cannot open credentials"))
	}
	auth, err := api.NewAuthenticator(cred)
	if err != nil {
		return nil, configurationError(errors.Wrap(err, "invalid credentials"))
	}

	token, err := os.Open("token.json")
	if err != nil {
		return nil, configurationError(errors.Wrap(err, "missing or invalid cached token"))
	}

	return auth.API(context.Background(), token)
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
