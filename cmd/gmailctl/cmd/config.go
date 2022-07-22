package cmd

import (
	"fmt"
	"os"
	"path"

	papply "github.com/mbrt/gmailctl/internal/engine/apply"
	"github.com/mbrt/gmailctl/internal/engine/cfgtest"
	"github.com/mbrt/gmailctl/internal/engine/config"
	"github.com/mbrt/gmailctl/internal/engine/config/v1alpha3"
	"github.com/mbrt/gmailctl/internal/errors"
)

type parseResult struct {
	Config v1alpha3.Config
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
		ts, err := cfgtest.NewFromParserRules(res.Res.Rules)
		if err != nil {
			stderrPrintf("WARNING: %d filters are excluded from the tests:\n", len(errors.Errors(err)))
			stderrPrintf("%+v\n", err)
		}
		tres := ts.ExecTests(res.Config.Tests)
		if !tres.OK {
			stderrPrintf("Test results: %s\n", tres)
			return res, fmt.Errorf("%d/%d config tests failed", len(tres.Failed), tres.NumTests)
		}
	}

	return res, err
}
