package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/mbrt/gmailctl/pkg/export/xml"
)

var (
	exportFilename  string
	exportOutput    string
	exportSkipTests bool
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export filters into the Gmail XML format",
	Long: `Export exports filters into the native Gmail XML format.
This allows to import them from within the Gmail settings or to share
them with other people.

By default export uses the configuration file inside the config
directory [config.(yaml|jsonnet)].`,
	Run: func(cmd *cobra.Command, args []string) {
		f := exportFilename
		if f == "" {
			f = configFilenameFromDir(cfgDir)
		}
		if err := export(f, exportOutput, !exportSkipTests); err != nil {
			fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	// Flags and configuration settings
	exportCmd.PersistentFlags().StringVarP(&exportFilename, "filename", "f", "", "configuration file")
	exportCmd.PersistentFlags().StringVarP(&exportOutput, "output", "o", "", "output file (defaut to stdout)")
	exportCmd.Flags().BoolVarP(&exportSkipTests, "yolo", "", false, "skip configuration tests")
}

func export(inputPath, outputPath string, test bool) (err error) {
	var out io.Writer
	if outputPath == "" {
		out = os.Stdout
	} else {
		f, e := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if e != nil {
			return fmt.Errorf("opening output: %w", err)
		}
		defer func() {
			e = f.Close()
			// do not hide more important error
			if err == nil {
				err = e
			}
		}()
		out = f
	}
	return exportWithOut(inputPath, out, test)
}

func exportWithOut(path string, out io.Writer, test bool) error {
	pres, err := parseConfig(path, "", test)
	if err != nil {
		return err
	}
	return xml.DefaultExporter().Export(pres.Config.Author, pres.Res.Filters, out)
}
