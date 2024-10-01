package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/mbrt/gmailctl/internal/engine/filter"
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
directory config.jsonnet].`,
	Run: func(*cobra.Command, []string) {
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

		fmt.Printf("# Search: %s\n", criteria.ToGmailSearch())
		fmt.Printf("# URL: %s\n", criteria.ToGmailSearchURL())
		cfg := parsedRules[i]
		b, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("marshalling rule: %w", err)
		}
		fmt.Println(string(b))
	}

	return nil
}
