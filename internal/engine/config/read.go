package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/google/go-jsonnet"

	"github.com/mbrt/gmailctl/internal/engine/config/v1alpha3"
	"github.com/mbrt/gmailctl/internal/errors"
)

const (
	// LatestVersion points to the latest version of the config format.
	LatestVersion = v1alpha3.Version

	unsupportedHelp = "Please see https://github.com/mbrt/gmailctl#yaml-config-is-unsupported.\n"
)

// ErrNotFound is returned when a file was not found.
var ErrNotFound = errors.New("config not found")

// ReadFile takes a path and returns the parsed config file.
//
// If the config file needs to have access to additional libraries,
// their location can be specified with cfgDirs.
func ReadFile(path, libPath string) (v1alpha3.Config, error) {
	/* #nosec */
	b, err := os.ReadFile(path)
	if err != nil {
		return v1alpha3.Config{}, errors.WithCause(err, ErrNotFound)
	}
	if ext := filepath.Ext(path); ext == ".yml" || ext == ".yaml" {
		return v1alpha3.Config{}, errors.WithDetails(errors.New("YAML config is unsupported"),
			unsupportedHelp)
	}
	// We pass the libPath to jsonnet, because that is the hint
	// to the libraries location. If no library is specified,
	// we use the original file location.
	if libPath == "" {
		libPath = path
	}
	return ReadJsonnet(libPath, b)
}

// ReadJsonnet parses a buffer containing a jsonnet config.
//
// The path is used to resolve imports.
func ReadJsonnet(p string, buf []byte) (v1alpha3.Config, error) {
	var res v1alpha3.Config
	vm := jsonnet.MakeVM()
	vm.Importer(&jsonnet.FileImporter{
		JPaths: []string{path.Dir(p)},
	})
	jstr, err := vm.EvaluateAnonymousSnippet(p, string(buf))
	if err != nil {
		return res, fmt.Errorf("parsing jsonnet: %w", err)
	}
	version, err := readJSONVersion(jstr)
	if err != nil {
		return res, fmt.Errorf("parsing the config version: %w", err)
	}
	if version != LatestVersion {
		return res, errors.WithDetails(fmt.Errorf("unsupported config version: %s", version),
			unsupportedHelp)
	}
	err = jsonUnmarshalStrict([]byte(jstr), &res)
	return res, err
}

func readJSONVersion(js string) (string, error) {
	// Try to unmarshal only the version
	v := struct {
		Version string `json:"version"`
	}{}
	err := json.Unmarshal([]byte(js), &v)
	return v.Version, err
}

func jsonUnmarshalStrict(buf []byte, v interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(buf))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		// Make the error more informative.
		jctx := contextFromJSONErr(err, buf)
		if jctx == "" {
			return err
		}
		return errors.WithDetails(err,
			fmt.Sprintf("JSON context:\n%s", jctx))
	}
	return nil
}

func contextFromJSONErr(err error, buf []byte) string {
	var (
		jserr  *json.SyntaxError
		juerr  *json.UnmarshalTypeError
		offset int
	)
	switch {
	case errors.As(err, &jserr):
		offset = int(jserr.Offset)
	case errors.As(err, &juerr):
		offset = int(juerr.Offset)
	default:
		return ""
	}

	if offset < 0 || offset >= len(buf) {
		return ""
	}

	// Collect 6 lines of context
	begin, end, count := 0, 0, 0
	for i := offset; i >= 0 && count < 3; i-- {
		if buf[i] == '\n' {
			begin = i + 1
			count++
		}
	}
	for i := offset; i < len(buf) && count < 6; i++ {
		if buf[i] == '\n' {
			end = i
			count++
		}
	}
	return string(buf[begin:end])
}
