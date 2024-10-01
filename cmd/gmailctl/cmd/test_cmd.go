package cmd

import "github.com/spf13/cobra"

var testFilename string

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Execute config tests",
	Long: `Execute config tests.
If the configuration file contains tests, they will be executed by
this command. Not all filters constructs are supported; if any
is found, the tests will ignore the entire filter and continue.
This might result in imprecise testing.

Warning: This command is still experimental.

List of unsupported constructs:
* Escaped expressions (pkg/config/v1alpha3/FilterNode.IsEscaped);
* Raw queries (pkg/config/v1alpha3/FilterNode.Query).

By default test uses the configuration file inside the config
directory [config.jsonnet].`,
	Run: func(*cobra.Command, []string) {
		f := testFilename
		if f == "" {
			f = configFilenameFromDir(cfgDir)
		}
		if err := test(f); err != nil {
			fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	// Flags and configuration settings
	testCmd.PersistentFlags().StringVarP(&testFilename, "filename", "f", "", "configuration file")
}

func test(path string) error {
	_, err := parseConfig(path, "", true)
	return err
}
