package cmd

import (
	"github.com/pkg/errors"

	"github.com/mbrt/gmailctl/pkg/config"
	cfgv2 "github.com/mbrt/gmailctl/pkg/config/v1alpha2"
	"github.com/mbrt/gmailctl/pkg/filter"
	"github.com/mbrt/gmailctl/pkg/parser"
)

type parseResult struct {
	config  cfgv2.Config
	rules   []parser.Rule
	filters filter.Filters
}

func parseConfig(path string) (parseResult, error) {
	var res parseResult
	var err error
	res.config, err = config.ReadFile(path)
	if err != nil {
		if config.IsNotFound(err) {
			return res, configurationError(err)
		}
		return res, errors.Wrap(err, "syntax error in config file")
	}

	if res.config.Version != config.LatestVersion {
		stderrPrintf("WARNING: Config file version '%s' is deprecated.\n",
			res.config.Version)
		stderrPrintf("  Please consider upgrading to version '%s'.\n\n",
			config.LatestVersion)
	}

	res.rules, err = parser.Parse(res.config)
	if err != nil {
		return res, errors.Wrap(err, "cannot parse config file")
	}

	res.filters, err = filter.FromRules(res.rules)
	if err != nil {
		return res, errors.Wrap(err, "error exporting to filters")
	}
	return res, nil
}
