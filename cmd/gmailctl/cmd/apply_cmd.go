package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	papply "github.com/mbrt/gmailctl/internal/engine/apply"
	"github.com/mbrt/gmailctl/internal/errors"
)

var (
	applyFilename     string
	applyYes          bool
	applyRemoveLabels bool
	applySkipTests    bool
	applyDebug        bool
	applyDiffContext  int
)

const renameLabelWarning = `Warning: You are going to delete labels. This operation is
irreversible, because it also removes those labels from messages.

If you are looking for renaming labels, please use the GMail UI.

`

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a configuration file to Gmail settings",
	Long: `The apply command applies minimal changes to your Gmail settings
to make them match your local configuration file.

By default apply uses the configuration file inside the config
directory [config.jsonnet].`,
	Run: func(*cobra.Command, []string) {
		f := applyFilename
		if f == "" {
			f = configFilenameFromDir(cfgDir)
		}
		if err := apply(f, !applyYes, !applySkipTests); err != nil {
			fatal(err)
		}
	}}

func init() {
	rootCmd.AddCommand(applyCmd)

	// Flags and configuration settings
	applyCmd.PersistentFlags().StringVarP(&applyFilename, "filename", "f", "", "configuration file")
	applyCmd.Flags().BoolVarP(&applyYes, "yes", "y", false, "don't ask for confirmation, just apply")
	applyCmd.Flags().BoolVarP(&applyRemoveLabels, "remove-labels", "r", false, "allow removing labels")
	applyCmd.Flags().BoolVarP(&applySkipTests, "yolo", "", false, "skip configuration tests")
	applyCmd.PersistentFlags().BoolVarP(&applyDebug, "debug", "", false, "print extra debugging information")
	applyCmd.PersistentFlags().IntVarP(&applyDiffContext, "diff-context", "", papply.DefaultContextLines, "number of lines of filter diff context to show")
}

func apply(path string, interactive, test bool) error {
	if applyDiffContext < 0 {
		return errors.New("--diff-context must be non-negative")
	}

	parseRes, err := parseConfig(path, "", test)
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

	diff, err := papply.Diff(parseRes.Res.GmailConfig, upstream, applyDebug, applyDiffContext)
	if err != nil {
		return fmt.Errorf("cannot compare upstream with local config: %w", err)
	}

	if diff.Empty() {
		fmt.Println("No changes have been made.")
		return nil
	}

	fmt.Printf("You are going to apply the following changes to your settings:\n\n%s\n", diff)

	if err := diff.Validate(); err != nil {
		return err
	}

	if len(diff.LabelsDiff.Removed) > 0 {
		fmt.Print(renameLabelWarning)
		if !applyRemoveLabels {
			return errors.WithDetails(errors.New("no changes have been made"),
				"To protect you, deletion is disabled unless you\n"+
					"explicitly provide the --remove-labels flag.\n")
		}
	}

	if interactive && !askYN("Do you want to apply them?") {
		return nil
	}

	fmt.Println("Applying the changes...")
	return papply.Apply(diff, gmailapi, applyRemoveLabels)
}

func configurationError(err error) error {
	return errors.WithDetails(err, "The configuration can be initialized with 'gmailctl init'")
}
