package cmd

import (
	"os"
	"path"

	"github.com/pkg/errors"

	papply "github.com/mbrt/gmailctl/pkg/apply"
	"github.com/mbrt/gmailctl/pkg/config"
	cfgv3 "github.com/mbrt/gmailctl/pkg/config/v1alpha3"
)

type parseResult struct {
	Config cfgv3.Config
	Res    papply.ConfigParseRes
}

func configFilenameFromDir(cfgDir string) string {
	f := path.Join(cfgDir, "config.yaml")
	if stat, err := os.Stat(f); err == nil && !stat.IsDir() {
		return f
	}
	return path.Join(cfgDir, "config.jsonnet")
}

func parseConfig(path, originalPath string) (parseResult, error) {
	var res parseResult
	var err error

	res.Config, err = config.ReadFile(path, originalPath)
	if err != nil {
		if config.IsNotFound(err) {
			return res, configurationError(err)
		}
		return res, errors.Wrap(err, "syntax error in config file")
	}

	if res.Config.Version != config.LatestVersion {
		stderrPrintf("WARNING: Config file version '%s' is deprecated.\n",
			res.Config.Version)
		stderrPrintf("  Please consider upgrading to version '%s'.\n\n",
			config.LatestVersion)
	}

	res.Res, err = papply.FromConfig(res.Config)
	return res, err
}
