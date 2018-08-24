package cmd

import (
	"io"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/mbrt/gmailfilter/pkg/config"
	"github.com/mbrt/gmailfilter/pkg/export/xml"
	"github.com/mbrt/gmailfilter/pkg/filter"
)

var (
	configFilename string
	configOutput   string
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export filters into the Gmail XML format",
	Long: `Export exports filters into the native Gmail XML format.
This allows to import them from within the Gmail settings or to share
them with other people.
`,
	Run: func(cmd *cobra.Command, args []string) {
		f := path.Join(cfgDir, "config.yaml")
		if configFilename != "" {
			f = configFilename
		}
		if err := export(f, configOutput); err != nil {
			fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	// Flags and configuration settings
	exportCmd.PersistentFlags().StringVarP(&configFilename, "filename", "f", "", "configuration file")
	exportCmd.PersistentFlags().StringVarP(&configOutput, "output", "o", "", "output file (defaut to stdout)")
}

func export(inputPath, outputPath string) (err error) {
	var out io.Writer
	if outputPath == "" {
		out = os.Stdout
	} else {
		f, e := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if e != nil {
			return errors.Wrap(err, "cannot open output")
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
	return exportWithOut(inputPath, out)
}

func exportWithOut(cfgPath string, out io.Writer) error {
	cfg, err := config.ParseFile(cfgPath)
	if err != nil {
		return errors.Wrap(err, "cannot parse config file")
	}

	filters, err := filter.FromConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "error exporting local filters")
	}
	return xml.DefaultExporter().Export(cfg.Author, filters, out)
}
