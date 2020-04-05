package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mbrt/gmailctl/pkg/api"
	papply "github.com/mbrt/gmailctl/pkg/apply"
	"github.com/mbrt/gmailctl/pkg/config"
)

// Parameters
var (
	editFilename  string
	editSkipTests bool
)

var (
	defaultEditors = []string{
		"editor",
		"nano",
		"vim",
		"vi",
	}

	errAbort     = errors.New("edit aborted")
	errUnchanged = errors.New("unchanged")
	errRetry     = errors.New("retry")
)

const abortHelp = `The original configuration is unchanged.
A temporary backup of your configuration has been saved at: %s`

// editCmd represents the apply command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit the configuration and apply it to Gmail",
	Long: `The edit command is a shortcut that allows you to edit the
configuration file, shows you the diff with your current Gmail
configuration, and applies minimal changes to it in order to
make it match your desired state.

The editor to be used can be overridden with the $EDITOR
environment variable.

By default edit uses the configuration file inside the config
directory [config.(yaml|jsonnet)].`,
	Run: func(cmd *cobra.Command, args []string) {
		f := editFilename
		if f == "" {
			f = configFilenameFromDir(cfgDir)
		}
		if err := edit(f, !editSkipTests); err != nil {
			fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	// Flags and configuration settings
	editCmd.PersistentFlags().StringVarP(&editFilename, "filename", "f", "", "configuration file")
	editCmd.Flags().BoolVarP(&editSkipTests, "yolo", "", false, "skip configuration tests")
}

func edit(path string, test bool) error {
	// First make sure that Gmail can be contacted, so that we don't
	// waste the user's time editing a config file that cannot be
	// applied now.
	gmailapi, err := openAPI()
	if err != nil {
		return configurationError(fmt.Errorf("connecting to Gmail: %w", err))
	}

	// Copy the configuration in a temporary file and edit it.
	tmpPath, err := copyToTmp(path)
	if err != nil {
		return err
	}

	for {
		if err = spawnEditor(tmpPath); err != nil {
			// Don't retry if the editor was aborted.
			// Try to cleanup the file
			_ = os.Remove(tmpPath)
			return err
		}
		if err = applyEdited(tmpPath, path, test, gmailapi); err != nil {
			if errors.Is(err, errUnchanged) {
				// Unchanged, but move the file anyways (it could be a refactoring)
				return moveFile(tmpPath, path)
			}
			if errors.Is(err, errAbort) {
				return UserError(err, fmt.Sprintf(abortHelp, tmpPath))
			}
			if errors.Is(err, errRetry) {
				continue
			}

			stderrPrintf("Error applying configuration: %v\n", err)
			if !askYN("Do you want to continue editing?") {
				return UserError(errAbort, fmt.Sprintf(abortHelp, tmpPath))
			}
			// Retry
			continue
		}

		// All good
		// Swap the configuration files.
		return moveFile(tmpPath, path)
	}
}

func moveFile(from, to string) error {
	// Swap the configuration files. Since these two can be in different
	// filesystems, we need to rewrite the file, instead of a simple rename.
	b, err := ioutil.ReadFile(from)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(to, os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if err != nil {
		return err
	}
	return os.Remove(from)
}

func copyToTmp(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", config.NotFoundError(err)
	}

	// Use the same extension as the original file (yaml | jsonnet)
	tmp, err := ioutil.TempFile("", fmt.Sprintf("gmailctl-*%s", filepath.Ext(path)))
	if err != nil {
		return "", fmt.Errorf("creating tmp file: %w", err)
	}

	if _, err := tmp.Write(b); err != nil {
		return "", err
	}

	res := tmp.Name()
	return res, tmp.Close()
}

func spawnEditor(path string) error {
	var editors []string
	if edvar := os.Getenv("EDITOR"); edvar != "" {
		editors = []string{edvar}
	}
	editors = append(editors, defaultEditors...)

	for _, editor := range editors {
		// $EDITOR may contain arguments, so we need to split
		// them away from the actual editor command.
		cmdargs := append(strings.Split(editor, " "), path)
		// #nosec
		cmd := exec.Command(cmdargs[0], cmdargs[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err == nil {
			return nil
		}
		if _, ok := err.(*exec.ExitError); ok {
			return errAbort
		}
	}

	return errors.New("no suitable editor found")
}

func applyEdited(path, originalPath string, test bool, gmailapi *api.GmailAPI) error {
	parseRes, err := parseConfig(path, originalPath, test)
	if err != nil {
		return err
	}

	upstream, err := upstreamConfig(gmailapi)
	if err != nil {
		return err
	}

	diff, err := papply.Diff(parseRes.Res.GmailConfig, upstream)
	if err != nil {
		return errors.New("comparing upstream with local config")
	}

	if diff.Empty() {
		fmt.Println("No changes have been made.")
		return nil
	}

	fmt.Printf("You are going to apply the following changes to your settings:\n\n%s\n", diff)

	if err := diff.Validate(); err != nil {
		return err
	}

	yesOption := "yes"
	if len(diff.LabelsDiff.Removed) > 0 {
		fmt.Print(renameLabelWarning)
		yesOption = "yes, and I ALSO WANT TO DELETE LABELS"
	}

	switch askOptions("Do you want to apply them?", []string{yesOption, "no (continue editing)", "abort"}) {
	case 0:
		break
	case 1:
		return errRetry
	default:
		return errAbort
	}

	fmt.Println("Applying the changes...")
	return papply.Apply(diff, gmailapi, true)
}
