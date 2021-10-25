package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"

	"github.com/google/go-jsonnet"
	"gopkg.in/yaml.v2"

	cfgv1 "github.com/mbrt/gmailctl/pkg/config/v1alpha1"
	cfgv2 "github.com/mbrt/gmailctl/pkg/config/v1alpha2"
	cfgv3 "github.com/mbrt/gmailctl/pkg/config/v1alpha3"
	"github.com/mbrt/gmailctl/pkg/errors"
)

// LatestVersion points to the latest version of the config format.
const LatestVersion = cfgv3.Version

// ErrNotFound is returned when a file was not found.
var ErrNotFound = errors.New("config not found")

// ReadFile takes a path and returns the parsed config file.
//
// If the config file needs to have access to additional libraries,
// their location can be specified with cfgDirs.
func ReadFile(path, libPath string) (cfgv3.Config, error) {
	/* #nosec */
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return cfgv3.Config{}, errors.WithCause(err, ErrNotFound)
	}
	if filepath.Ext(path) == ".jsonnet" {
		// We pass the libPath to jsonnet, because that is the hint
		// to the libraries location. If no library is specified,
		// we use the original file location.
		if libPath != "" {
			return readJsonnet(libPath, b)
		}
		return readJsonnet(path, b)
	}
	return readYaml(b)
}

func readJsonnet(p string, buf []byte) (cfgv3.Config, error) {
	var res cfgv3.Config
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

	switch version {
	case cfgv3.Version:
		err = jsonUnmarshalStrict([]byte(jstr), &res)
		return res, err

	case cfgv2.Version:
		var v2 cfgv2.Config
		err = jsonUnmarshalStrict([]byte(jstr), &v2)
		if err != nil {
			return res, fmt.Errorf("parsing v1alpha2 config: %w", err)
		}
		return importFromV2(v2)

	case cfgv1.Version:
		var v1 cfgv1.Config
		err = jsonUnmarshalStrict([]byte(jstr), &v1)
		if err != nil {
			return res, fmt.Errorf("parsing v1alpha1 config: %w", err)
		}
		return importFromV1(v1)

	default:
		return res, fmt.Errorf("unknown config version: %s", version)
	}
}

func readYaml(buf []byte) (cfgv3.Config, error) {
	var res cfgv3.Config
	version, err := readYamlVersion(buf)
	if err != nil {
		return res, fmt.Errorf("parsing the config version: %w", err)
	}

	switch version {
	case cfgv2.Version:
		var v2 cfgv2.Config
		err = yaml.UnmarshalStrict(buf, &v2)
		if err != nil {
			return res, fmt.Errorf("parsing v1alpha2 config: %w", err)
		}
		return importFromV2(v2)

	case cfgv1.Version:
		var v1 cfgv1.Config
		err = yaml.UnmarshalStrict(buf, &v1)
		if err != nil {
			return res, fmt.Errorf("parsing v1alpha1 config: %w", err)
		}
		return importFromV1(v1)

	default:
		return res, fmt.Errorf("unsupported config version: %s", version)
	}
}

func importFromV1(v1 cfgv1.Config) (cfgv3.Config, error) {
	v2, err := cfgv2.Import(v1)
	if err != nil {
		return cfgv3.Config{}, err
	}
	return importFromV2(v2)
}

func importFromV2(v2 cfgv2.Config) (cfgv3.Config, error) {
	return cfgv3.Import(v2)
}

func readYamlVersion(buf []byte) (string, error) {
	// Try to unmarshal only the version
	v := struct {
		Version string `yaml:"version"`
	}{}
	err := yaml.Unmarshal(buf, &v)
	return v.Version, err
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
