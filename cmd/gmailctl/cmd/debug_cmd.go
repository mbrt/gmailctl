package cmd

import (
	"fmt"
	"net/url"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/mbrt/gmailctl/pkg/filter"
)

var (
	debugFilename string
)

// debugCmd represents the debug command
var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Shows an annotated version of the configuration",
	Long: `The debug command shows an annotated version of the configuration
with handy URLs to Gmail search that can be used to test that the
filter applies to the intended emails.

By default debug uses the "config.yaml" file inside the config directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		f := path.Join(cfgDir, "config.yaml")
		if debugFilename != "" {
			f = debugFilename
		}
		if err := debug(f); err != nil {
			fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)

	// Flags and configuration settings
	debugCmd.PersistentFlags().StringVarP(&debugFilename, "filename", "f", "", "configuration file")
}

func debug(path string) error {
	parseRes, err := parseConfig(path)
	if err != nil {
		return err
	}

	if exp, got := len(parseRes.config.Rules), len(parseRes.rules); exp != got {
		return errors.Errorf(
			"unexpected number of generated rules: got %d, expected %d",
			got, exp,
		)
	}

	for i := 0; i < len(parseRes.rules); i++ {
		parsed := parseRes.rules[i]
		criteria, err := filter.GenerateCriteria(parsed.Criteria)
		if err != nil {
			return errors.Wrap(err, "error generating criteria")
		}
		search := criteria.ToGmailSearch()

		fmt.Printf("# Search: %s\n", search)
		fmt.Printf("# URL: %s\n", toGmailURL(search))
		cfg := parseRes.config.Rules[i]
		b, err := yaml.Marshal(cfg)
		if err != nil {
			return errors.Wrap(err, "error marshalling rule")
		}
		fmt.Println(string(b))
	}

	return nil
}

func toGmailURL(s string) string {
	return fmt.Sprintf(
		"https://mail.google.com/mail/u/0/#search/%s",
		url.PathEscape(s),
	)
}
