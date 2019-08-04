package apply

import (
	cfgv3 "github.com/mbrt/gmailctl/pkg/config/v1alpha3"
	"github.com/mbrt/gmailctl/pkg/filter"
	"github.com/mbrt/gmailctl/pkg/label"
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

// Diff computes the diff between local and upstream configuration.
func Diff(cfg cfgv3.Config, upstream GmailConfig) (ConfigDiff, error) {
	// TODO
	return ConfigDiff{}, nil
}

// Apply applies the changes identified by the diff to the remote configuration.
func Apply(d ConfigDiff, api interface{}) error {
	// TODO
	return nil
}
