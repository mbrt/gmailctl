package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"

	papply "github.com/mbrt/gmailctl/pkg/apply"
	"github.com/mbrt/gmailctl/pkg/cfgtest"
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

func parseConfig(path, originalPath string, test bool) (parseResult, error) {
	var res parseResult
	var err error

	res.Config, err = config.ReadFile(path, originalPath)
	if err != nil {
		if errors.Is(err, config.ErrNotFound) {
			return res, configurationError(err)
		}
		return res, fmt.Errorf("syntax error in config file: %w", err)
	}

	if res.Config.Version != config.LatestVersion {
		stderrPrintf("WARNING: Config file version '%s' is deprecated.\n",
			res.Config.Version)
		stderrPrintf("  Please consider upgrading to version '%s'.\n\n",
			config.LatestVersion)
	}

	res.Res, err = papply.FromConfig(res.Config)
	if err != nil {
		return res, err
	}
	if test && len(res.Config.Tests) > 0 {
		ts, errs := cfgtest.NewFromParserRules(res.Res.Rules)
		if len(errs) > 0 {
			stderrPrintf("WARNING: %d filters are excluded from the tests:\n", len(errs))
			for _, err := range errs {
				stderrPrintf("  %v\n", err)
			}
		}
		err = ts.ExecTests(res.Config.Tests)
		if err != nil {
			return res, fmt.Errorf("config tests failed: %w", err)
		}
	}

	return res, err
}
