package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/mbrt/gmailctl/pkg/config"
	"github.com/mbrt/gmailctl/pkg/config/v1alpha1"
	"github.com/mbrt/gmailctl/pkg/config/v1alpha2"
	"github.com/mbrt/gmailctl/pkg/config/v1alpha3"
)

func importConfig(path string) (string, error) {
	cfg, err := readConfig(path)
	if err != nil {
		return "", err
	}
	if cfg.Version == config.LatestVersion {
		return "", fmt.Errorf("config version %s is already up-to-date", cfg.Version)
	}
	return "TODO", nil
}

func readConfig(path string) (v1alpha3.Config, error) {
	/* #nosec */
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return v1alpha3.Config{}, err
	}
	return readYaml(buf)
}

func readYaml(buf []byte) (v1alpha3.Config, error) {
	var res v1alpha3.Config
	version, err := readYamlVersion(buf)
	if err != nil {
		return res, fmt.Errorf("parsing the config version: %w", err)
	}

	switch version {
	case v1alpha3.Version:
		var v3 v1alpha3.Config
		err = yaml.Unmarshal(buf, &v3)
		if err != nil {
			return res, fmt.Errorf("parsing the v1alpha3 config: %w", err)
		}
		return v3, nil

	case v1alpha2.Version:
		var v2 v1alpha2.Config
		err = yaml.UnmarshalStrict(buf, &v2)
		if err != nil {
			return res, fmt.Errorf("parsing v1alpha2 config: %w", err)
		}
		return importFromV2(v2)

	case v1alpha1.Version:
		var v1 v1alpha1.Config
		err = yaml.UnmarshalStrict(buf, &v1)
		if err != nil {
			return res, fmt.Errorf("parsing v1alpha1 config: %w", err)
		}
		return importFromV1(v1)

	default:
		return res, fmt.Errorf("unsupported config version: %s", version)
	}
}

func importFromV1(v1 v1alpha1.Config) (v1alpha3.Config, error) {
	v2, err := v1alpha2.Import(v1)
	if err != nil {
		return v1alpha3.Config{}, err
	}
	return importFromV2(v2)
}

func importFromV2(v2 v1alpha2.Config) (v1alpha3.Config, error) {
	return v1alpha3.Import(v2)
}

func readYamlVersion(buf []byte) (string, error) {
	// Try to unmarshal only the version
	v := struct {
		Version string `yaml:"version"`
	}{}
	err := yaml.Unmarshal(buf, &v)
	return v.Version, err
}
