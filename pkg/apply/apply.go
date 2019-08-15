package apply

import (
	"sort"
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

// Empty returns whether the diff contains no changes.
func (d ConfigDiff) Empty() bool {
	return d.Filters.Empty() && d.Labels.Empty()
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

// API provides access to Gmail APIs.
type API interface {
	AddLabels(lbs label.Labels) error
	AddFilters(fs filter.Filters) error
	UpdateLabels(lbs label.Labels) error
	DeleteFilters(ids []string) error
	DeleteLabels(ids []string) error
}

// Apply applies the changes identified by the diff to the remote configuration.
func Apply(d ConfigDiff, api API) error {
	// In order to prevent not found errors, the sequence has to be:
	//
	// - add new labels
	// - add new filters
	// - modify labels
	// - remove filters
	// - remove labels

	if err := addLabels(d.Labels.Added, api); err != nil {
		return errors.Wrap(err, "error creating labels")
	}
	if err := addFilters(d.Filters.Added, api); err != nil {
		return errors.Wrap(err, "error creating filters")
	}
	if err := updateLabels(d.Labels.Modified, api); err != nil {
		return errors.Wrap(err, "error updating labels")
	}
	if err := removeFilters(d.Filters.Removed, api); err != nil {
		return errors.Wrap(err, "error deleting filters")
	}
	if err := removeLabels(d.Labels.Removed, api); err != nil {
		return errors.Wrap(err, "error removing labels")
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
