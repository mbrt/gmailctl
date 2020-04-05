package apply

import (
	"fmt"
	"sort"
	"strings"

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

// ConfigParseRes represents the result of a config parse.
type ConfigParseRes struct {
	GmailConfig
	Rules []parser.Rule
}

// FromConfig creates a GmailConfig from a parsed configuration file.
func FromConfig(cfg cfgv3.Config) (ConfigParseRes, error) {
	res := ConfigParseRes{}
	var err error

	res.Rules, err = parser.Parse(cfg)
	if err != nil {
		return res, fmt.Errorf("cannot parse config file: %w", err)
	}
	res.Filters, err = filter.FromRules(res.Rules)
	if err != nil {
		return res, fmt.Errorf("exporting to filters: %w", err)
	}
	res.Labels = label.FromConfig(cfg.Labels)

	return res, nil
}

// ConfigDiff contains the difference between local and upstream configuration,
// including both labels and filters.
//
// For validation purposes, the local config is also kept.
type ConfigDiff struct {
	FiltersDiff filter.FiltersDiff
	LabelsDiff  label.LabelsDiff

	LocalConfig GmailConfig
}

func (d ConfigDiff) String() string {
	var res []string

	if !d.FiltersDiff.Empty() {
		res = append(res, "Filters:")
		res = append(res, d.FiltersDiff.String())
	}
	if !d.LabelsDiff.Empty() {
		res = append(res, "Labels:")
		res = append(res, d.LabelsDiff.String())
	}

	return strings.Join(res, "\n")
}

// Empty returns whether the diff contains no changes.
func (d ConfigDiff) Empty() bool {
	return d.FiltersDiff.Empty() && d.LabelsDiff.Empty()
}

// Validate returns whether the given diff is valid.
func (d ConfigDiff) Validate() error {
	if d.LabelsDiff.Empty() {
		return nil
	}
	if err := d.LocalConfig.Labels.Validate(); err != nil {
		return fmt.Errorf("validating labels: %w", err)
	}
	if err := label.Validate(d.LabelsDiff, d.LocalConfig.Filters); err != nil {
		return fmt.Errorf("invalid labels diff: %w", err)
	}
	return nil
}

// Diff computes the diff between local and upstream configuration.
func Diff(local, upstream GmailConfig) (ConfigDiff, error) {
	res := ConfigDiff{
		LocalConfig: local,
	}
	var err error

	res.FiltersDiff, err = filter.Diff(upstream.Filters, local.Filters)
	if err != nil {
		return res, fmt.Errorf("cannot compute filters diff: %w", err)
	}

	if len(local.Labels) > 0 {
		// LabelsDiff management opted-in
		res.LabelsDiff, err = label.Diff(upstream.Labels, local.Labels)
		if err != nil {
			return res, fmt.Errorf("cannot compute labels diff: %w", err)
		}
	}

	return res, nil
}

// API provides access to Gmail APIs.
type API interface {
	AddLabels(lbs label.Labels) error
	AddFilters(fs filter.Filters) error
	UpdateLabels(lbs label.Labels) error
	DeleteFilters(ids []string) error
	DeleteLabels(ids []string) error
}

// Apply applies the changes identified by the diff to the remote configuration.
func Apply(d ConfigDiff, api API, allowRemoveLabels bool) error {
	// In order to prevent not found errors, the sequence has to be:
	//
	// - add new labels
	// - add new filters
	// - modify labels
	// - remove filters
	// - remove labels

	if err := addLabels(d.LabelsDiff.Added, api); err != nil {
		return fmt.Errorf("creating labels: %w", err)
	}
	if err := addFilters(d.FiltersDiff.Added, api); err != nil {
		return fmt.Errorf("creating filters: %w", err)
	}
	if err := updateLabels(d.LabelsDiff.Modified, api); err != nil {
		return fmt.Errorf("updating labels: %w", err)
	}
	if err := removeFilters(d.FiltersDiff.Removed, api); err != nil {
		return fmt.Errorf("deleting filters: %w", err)
	}

	if !allowRemoveLabels {
		return nil
	}
	if err := removeLabels(d.LabelsDiff.Removed, api); err != nil {
		return fmt.Errorf("removing labels: %w", err)
	}

	return nil
}

func addLabels(lbs label.Labels, api API) error {
	if len(lbs) == 0 {
		return nil
	}
	// If we have nested labels we should create them in the right order.
	// As a quick hack, we could sort them by the length of the name,
	// because a label is strictly longer than its prefixes.
	sort.Sort(byLen(lbs))
	return api.AddLabels(lbs)
}

func addFilters(ls filter.Filters, api API) error {
	if len(ls) > 0 {
		return api.AddFilters(ls)
	}
	return nil
}

func updateLabels(ms []label.ModifiedLabel, api API) error {
	if len(ms) == 0 {
		return nil
	}
	var lbs label.Labels
	for _, m := range ms {
		label := m.New
		label.ID = m.Old.ID
		lbs = append(lbs, label)
	}
	return api.UpdateLabels(lbs)
}

func removeFilters(ls filter.Filters, api API) error {
	if len(ls) == 0 {
		return nil
	}
	ids := make([]string, len(ls))
	for i, f := range ls {
		ids[i] = f.ID
	}
	return api.DeleteFilters(ids)
}

func removeLabels(lbs label.Labels, api API) error {
	if len(lbs) == 0 {
		return nil
	}
	// If we have nested labels we should remove them in the right order.
	// As a quick hack, we could sort them by the length of the name,
	// because a label is strictly longer than its prefixes.
	sort.Sort(byLen(lbs))

	// Delete in reverse order
	var ids []string
	for i := len(lbs) - 1; i >= 0; i-- {
		ids = append(ids, lbs[i].ID)
	}
	return api.DeleteLabels(ids)
}

type byLen label.Labels

func (b byLen) Len() int {
	return len(b)
}

func (b byLen) Less(i, j int) bool {
	return len(b[i].Name) < len(b[j].Name)
}

func (b byLen) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
