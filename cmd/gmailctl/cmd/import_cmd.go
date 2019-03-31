package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import filters from Gmail to a local file",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := doImport(); err != nil {
			fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}

func doImport() error {
	gmailapi, err := openAPI()
	if err != nil {
		return configurationError(errors.Wrap(err, "cannot connect to Gmail"))
	}

	upstream, err := upstreamFilters(gmailapi)
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(upstream, "", "  ")
	if err != nil {
		return errors.Wrap(err, "error converting to JSON")
	}

	fmt.Println(string(b))
	return nil
}
