package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	pexport "github.com/mbrt/gmailctl/internal/engine/export"
	"github.com/mbrt/gmailctl/internal/engine/export/xml"
)

var (
	exportFilename    string
	exportOutput      string
	exportSkipTests   bool
	exportNew         bool
	exportChangesOnly bool
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export filters into the Gmail XML format",
	Long: `Export exports filters into the native Gmail XML format.
This allows to import them from within the Gmail settings or to share
them with other people.

By default export uses the configuration file inside the config
directory [config.jsonnet].

With the --new flag, only filters that are newly added or modified
compared to your current Gmail settings are exported.`,
	Run: func(*cobra.Command, []string) {
		f := exportFilename
		if f == "" {
			f = configFilenameFromDir(cfgDir)
		}
		if err := export(f, exportOutput, !exportSkipTests, exportNew || exportChangesOnly); err != nil {
			fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	// Flags and configuration settings
	exportCmd.PersistentFlags().StringVarP(&exportFilename, "filename", "f", "", "configuration file")
	exportCmd.PersistentFlags().StringVarP(&exportOutput, "output", "o", "", "output file (default to stdout)")
	exportCmd.Flags().BoolVarP(&exportSkipTests, "yolo", "", false, "skip configuration tests")
	exportCmd.Flags().BoolVar(&exportNew, "new", false, "export only newly added or modified rules")
	exportCmd.Flags().BoolVar(&exportChangesOnly, "changes-only", false, "alias for --new")
}

func export(inputPath, outputPath string, test, newOnly bool) (err error) {
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
	return exportWithOut(inputPath, out, test, newOnly)
}

func exportWithOut(path string, out io.Writer, test, newOnly bool) error {
	pres, err := parseConfig(path, "", test)
	if err != nil {
		return err
	}
	if newOnly {
		gmailapi, err := openAPI()
		if err != nil {
			return configurationError(fmt.Errorf("cannot connect to Gmail: %w", err))
		}
		upstream, err := upstreamConfig(gmailapi)
		if err != nil {
			return err
		}
		diffExporter := pexport.NewRuleDiffExporter()
		return diffExporter.ExportNew(pres.Config.Author, pres.Res.Filters, upstream.Filters, out)
	}
	return xml.DefaultExporter().Export(pres.Config.Author, pres.Res.Filters, out)
}
