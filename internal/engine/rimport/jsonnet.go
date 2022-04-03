package rimport

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"regexp"
)

const labelsComment = `  // Note: labels management is optional. If you prefer to use the
  // GMail interface to add and remove labels, you can safely remove
  // this section of the config.
`

var labelsLine = "  labels: ["

func MarshalJsonnet(v interface{}, w io.Writer, header string) error {
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

	_, err = writer.WriteString(header)
	if err != nil {
		return err
	}

	line, _, err = reader.ReadLine()
	for err == nil {
		line = replaceGroupsRe(keyRe, line)
		if string(line) == labelsLine {
			_, err = writer.WriteString(labelsComment)
			if err != nil {
				break
			}
		}
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

	if errors.Is(err, io.EOF) {
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
