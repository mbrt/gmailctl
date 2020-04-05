package cmd

import (
	"fmt"
	"net/url"

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

By default debug uses the configuration file inside the config
directory [config.(yaml|jsonnet)].`,
	Run: func(cmd *cobra.Command, args []string) {
		f := debugFilename
		if f == "" {
			f = configFilenameFromDir(cfgDir)
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
	parseRes, err := parseConfig(path, "", false)
	if err != nil {
		return err
	}
	configRules := parseRes.Config.Rules
	parsedRules := parseRes.Res.Rules

	if exp, got := len(configRules), len(parsedRules); exp != got {
		return fmt.Errorf(
			"unexpected number of generated rules: got %d, expected %d",
			got, exp,
		)
	}

	for i := 0; i < len(parsedRules); i++ {
		parsed := parsedRules[i]
		criteria, err := filter.GenerateCriteria(parsed.Criteria)
		if err != nil {
			return fmt.Errorf("generating criteria: %w", err)
		}
		search := criteria.ToGmailSearch()

		fmt.Printf("# Search: %s\n", search)
		fmt.Printf("# URL: %s\n", toGmailURL(search))
		cfg := parsedRules[i]
		b, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("marshalling rule: %w", err)
		}
		fmt.Println(string(b))
	}

	return nil
}

func toGmailURL(s string) string {
	return fmt.Sprintf(
		"https://mail.google.com/mail/u/0/#search/%s",
		url.QueryEscape(s),
	)
}
