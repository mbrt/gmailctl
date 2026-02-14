package apply

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/mbrt/gmailctl/internal/engine/config/v1alpha3"
	"github.com/mbrt/gmailctl/internal/engine/filter"
	"github.com/mbrt/gmailctl/internal/engine/label"
	"github.com/mbrt/gmailctl/internal/engine/parser"
	"github.com/schollz/progressbar/v3"
)

// DefaultContextLines is the default number of lines of context to show in the filter diff.
const DefaultContextLines = 5

// progress wraps a progressbar to encapsulate nil checks.
type progress struct {
	bar *progressbar.ProgressBar
}

func (p *progress) Describe(desc string) {
	if p.bar != nil {
		p.bar.Describe(desc)
	}
}

func (p *progress) Add(n int) {
	if p.bar != nil {
		if err := p.bar.Add(n); err != nil {
			fmt.Fprintf(os.Stderr, "progress bar update failed: %v\n", err)
		}
	}
}

func (p *progress) Finish() {
	if p.bar != nil {
		p.bar.Finish()
	}
}

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
func FromConfig(cfg v1alpha3.Config) (ConfigParseRes, error) {
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

// FetchAPI provides access to Gmail get APIs.
type FetchAPI interface {
	ListFilters() (filter.Filters, error)
	ListLabels() (label.Labels, error)
}

// FromConfig creates a GmailConfig from Gmail APIs.
func FromAPI(api FetchAPI) (GmailConfig, error) {
	l, err := api.ListLabels()
	if err != nil {
		return GmailConfig{}, fmt.Errorf("listing labels from Gmail: %w", err)
	}
	f, err := api.ListFilters()
	if err != nil {
		if len(f) == 0 {
			return GmailConfig{}, fmt.Errorf("getting filters from Gmail: %w", err)
		}
		// Some upstream filters may be invalid and in most cases we just want to ignore
		// those and carry on.
	}
	return GmailConfig{
		Labels:  l,
		Filters: f,
	}, err
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
func Diff(local, upstream GmailConfig, debugInfo bool, contextLines int, colorize bool) (ConfigDiff, error) {
	res := ConfigDiff{
		LocalConfig: local,
	}
	var err error

	res.FiltersDiff, err = filter.Diff(upstream.Filters, local.Filters, debugInfo, contextLines, colorize)
	if err != nil {
		return res, fmt.Errorf("cannot compute filters diff: %w", err)
	}

	if len(local.Labels) > 0 {
		// LabelsDiff management opted-in
		res.LabelsDiff, err = label.Diff(upstream.Labels, local.Labels, colorize)
		if err != nil {
			return res, fmt.Errorf("cannot compute labels diff: %w", err)
		}
	}

	return res, nil
}

// API provides access to Gmail APIs.
type API interface {
	AddLabels(lbs label.Labels, onProgress ...func()) error
	AddFilters(fs filter.Filters, onProgress ...func()) error
	UpdateLabels(lbs label.Labels, onProgress ...func()) error
	DeleteFilters(ids []string, onProgress ...func()) error
	DeleteLabels(ids []string, onProgress ...func()) error
}

// Apply applies the changes identified by the diff to the remote configuration.
func Apply(d ConfigDiff, api API, allowRemoveLabels bool, showProgress bool) error {
	// In order to prevent not found errors, the sequence has to be:
	//
	// - add new labels
	// - add new filters
	// - modify labels
	// - remove filters
	// - remove labels

	// Calculate total items to process
	total := len(d.LabelsDiff.Added) +
		len(d.FiltersDiff.Added) +
		len(d.LabelsDiff.Modified) +
		len(d.FiltersDiff.Removed)
	if allowRemoveLabels {
		total += len(d.LabelsDiff.Removed)
	}

	// Create progress wrapper (handles nil internally)
	bar := &progress{}
	if showProgress && total > 0 {
		bar.bar = progressbar.NewOptions(total,
			progressbar.OptionSetDescription("Applying changes"),
			progressbar.OptionShowCount(),
			progressbar.OptionSetWriter(os.Stderr),
		)
	}
	defer bar.Finish()

	if err := addLabels(d.LabelsDiff.Added, api, bar); err != nil {
		return fmt.Errorf("creating labels: %w", err)
	}
	if err := addFilters(d.FiltersDiff.Added, api, bar); err != nil {
		return fmt.Errorf("creating filters: %w", err)
	}
	if err := updateLabels(d.LabelsDiff.Modified, api, bar); err != nil {
		return fmt.Errorf("updating labels: %w", err)
	}
	if err := removeFilters(d.FiltersDiff.Removed, api, bar); err != nil {
		return fmt.Errorf("deleting filters: %w", err)
	}

	if !allowRemoveLabels {
		return nil
	}
	if err := removeLabels(d.LabelsDiff.Removed, api, bar); err != nil {
		return fmt.Errorf("removing labels: %w", err)
	}

	return nil
}

func addLabels(lbs label.Labels, api API, bar *progress) error {
	if len(lbs) == 0 {
		return nil
	}
	// If we have nested labels we should create them in the right order.
	// As a quick hack, we could sort them by the length of the name,
	// because a label is strictly longer than its prefixes.
	sort.Sort(byLen(lbs))

	bar.Describe("Adding labels")

	if err := api.AddLabels(lbs, func() { bar.Add(1) }); err != nil {
		return err
	}

	return nil
}

func addFilters(fs filter.Filters, api API, bar *progress) error {
	if len(fs) == 0 {
		return nil
	}

	bar.Describe("Adding filters")

	if err := api.AddFilters(fs, func() { bar.Add(1) }); err != nil {
		return err
	}

	return nil
}

func updateLabels(ms []label.ModifiedLabel, api API, bar *progress) error {
	if len(ms) == 0 {
		return nil
	}

	bar.Describe("Updating labels")

	// Prepare all labels with their IDs
	lbs := make(label.Labels, len(ms))
	for i, m := range ms {
		lb := m.New
		lb.ID = m.Old.ID
		lbs[i] = lb
	}

	if err := api.UpdateLabels(lbs, func() { bar.Add(1) }); err != nil {
		return err
	}

	return nil
}

func removeFilters(fs filter.Filters, api API, bar *progress) error {
	if len(fs) == 0 {
		return nil
	}

	bar.Describe("Removing filters")

	ids := make([]string, len(fs))
	for i, f := range fs {
		ids[i] = f.ID
	}

	if err := api.DeleteFilters(ids, func() { bar.Add(1) }); err != nil {
		return err
	}

	return nil
}

func removeLabels(lbs label.Labels, api API, bar *progress) error {
	if len(lbs) == 0 {
		return nil
	}
	// If we have nested labels we should remove them in the right order.
	// As a quick hack, we could sort them by the length of the name,
	// because a label is strictly longer than its prefixes.
	sort.Sort(byLen(lbs))

	bar.Describe("Removing labels")

	// Collect IDs in reverse order (longest names first = deepest nested first)
	ids := make([]string, len(lbs))
	for i := 0; i < len(lbs); i++ {
		ids[i] = lbs[len(lbs)-1-i].ID
	}

	if err := api.DeleteLabels(ids, func() { bar.Add(1) }); err != nil {
		return err
	}

	return nil
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
