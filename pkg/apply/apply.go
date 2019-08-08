package apply

import (
	"strings"

	"github.com/pkg/errors"

	cfgv3 "github.com/mbrt/gmailctl/pkg/config/v1alpha3"
	"github.com/mbrt/gmailctl/pkg/filter"
	"github.com/mbrt/gmailctl/pkg/label"
	"github.com/mbrt/gmailctl/pkg/parser"
)

// GmailConfig represents a Gmail configuration.
type GmailConfig struct {
	Labels  label.Labels
	Filters filter.Filters
}

// ConfigDiff contains the difference between local and upstream configuration,
// including both labels and filters.
type ConfigDiff struct {
	Filters filter.FiltersDiff
	Labels  label.LabelsDiff
}

func (d ConfigDiff) String() string {
	var res []string

	if !d.Filters.Empty() {
		res = append(res, "Filters:")
		res = append(res, d.Filters.String())
	}
	if !d.Labels.Empty() {
		res = append(res, "Labels:")
		res = append(res, d.Labels.String())
	}

	return strings.Join(res, "\n")
}

// Diff computes the diff between local and upstream configuration.
func Diff(cfg cfgv3.Config, upstream GmailConfig) (ConfigDiff, error) {
	rules, err := parser.Parse(cfg)
	if err != nil {
		return ConfigDiff{}, errors.Wrap(err, "cannot parse config file")
	}
	filters, err := filter.FromRules(rules)
	if err != nil {
		return ConfigDiff{}, errors.Wrap(err, "error exporting to filters")
	}
	fdiff, err := filter.Diff(upstream.Filters, filters)
	if err != nil {
		return ConfigDiff{}, errors.Wrap(err, "cannot compute filters diff")
	}

	ldiff := label.LabelsDiff{}
	if len(cfg.Labels) > 0 {
		// Labels management opted-in
		labels := label.FromConfig(cfg.Labels)
		if err = labels.Validate(); err != nil {
			return ConfigDiff{}, errors.Wrap(err, "error validating labels")
		}
		ldiff, err = label.Diff(upstream.Labels, labels)
		if err != nil {
			return ConfigDiff{}, errors.Wrap(err, "cannot compute labels diff")
		}
		if err = label.Validate(ldiff, filters); err != nil {
			return ConfigDiff{}, errors.Wrap(err, "invalid labels diff")
		}
	}

	return ConfigDiff{
		Filters: fdiff,
		Labels:  ldiff,
	}, nil
}

// Apply applies the changes identified by the diff to the remote configuration.
func Apply(d ConfigDiff, api interface{}) error {
	// TODO
	return nil
}
