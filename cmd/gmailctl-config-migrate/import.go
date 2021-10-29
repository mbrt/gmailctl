package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/mbrt/gmailctl/pkg/config"
	cfgv1 "github.com/mbrt/gmailctl/pkg/config/v1alpha1"
	cfgv2 "github.com/mbrt/gmailctl/pkg/config/v1alpha2"
	cfgv3 "github.com/mbrt/gmailctl/pkg/config/v1alpha3"
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

func readConfig(path string) (cfgv3.Config, error) {
	/* #nosec */
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return cfgv3.Config{}, err
	}
	if err != nil {
		return cfgv3.Config{}, err
	}
	return readYaml(buf)
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
