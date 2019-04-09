package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"regexp"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/mbrt/gmailctl/pkg/rimport"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import filters from Gmail to a local file",
	Long:  `WARNING: Experimental config generation from remote filters.`,
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

	cfg, err := rimport.Import(upstream)
	if err != nil {
		return err
	}

	err = marshalJsonnet(cfg, os.Stdout)
	if err != nil {
		return errors.Wrap(err, "error converting to Jsonnet")
	}
	return nil
}

func marshalJsonnet(v interface{}, w io.Writer) error {
	// Convert to JSON
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	// Make JSON resemble Jsonnet by getting rid of unnecessary quotes
	reader := bufio.NewReader(bytes.NewReader(b))
	writer := bufio.NewWriter(w)
	keyRe := regexp.MustCompile(`^ *"([a-zA-Z01]+)":`)
	var line []byte

	line, _, err = reader.ReadLine()
	for err == nil {
		line = replaceGroupsRe(keyRe, line)
		_, err = writer.Write(line)
		if err != nil {
			break
		}
		_, err = writer.WriteRune('\n')
		if err != nil {
			break
		}

		line, _, err = reader.ReadLine()
	}

	if err == io.EOF {
		return writer.Flush()
	}
	return err
}

func replaceGroupsRe(re *regexp.Regexp, in []byte) []byte {
	m := re.FindSubmatchIndex(in)
	if len(m) == 0 {
		return in
	}
	keyb, keye := m[2], m[3]

	var res []byte
	res = append(res, in[:keyb-1]...)
	res = append(res, in[keyb:keye]...)
	res = append(res, in[keye+1:]...)
	return res
}
