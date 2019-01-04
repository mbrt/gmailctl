package v2alpha1

import (
	"github.com/pkg/errors"

	v1 "github.com/mbrt/gmailctl/pkg/config/v1alpha1"
)

// Version is the latest configuration version.
const Version = "v1alpha1"

// Config contains the yaml configuration of the Gmail filters.
type Config struct {
	Version string        `yaml:"version"`
	Author  Author        `yaml:"author,omitempty"`
	Filters []NamedFilter `yaml:"filters,omitempty"`
	Rules   []Rule        `yaml:"rules"`
}

// Valid returns an error if the configuration is invalid.
func (c Config) Valid() error {
	if c.Version != Version {
		return errors.Errorf("invalid version: %s", c.Version)
	}

	var filterNames map[string]struct{}
	for _, f := range c.Filters {
		if err := f.Valid(filterNames); err != nil {
			return errors.Wrap(err, "invalid filter")
		}
		filterNames[f.Name] = struct{}{}
	}

	for _, r := range c.Rules {
		if err := r.Valid(filterNames); err != nil {
			return errors.Wrap(err, "invalid filter")
		}
	}
	return nil
}

// Author represents the owner of the gmail account.
type Author struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

type NamedFilter struct {
	Name  string `yaml:"name"`
	Query Filter `yaml:"filter"`
}

func (f NamedFilter) Valid(otherFilters map[string]struct{}) error {
	// TODO
	return nil
}

type Filter struct {
	And []Filter `yaml:"and,omitempty"`
	Or  []Filter `yaml:"or,omitempty"`
	Not *Filter  `yaml:"not,omitempty"`

	From    string `yaml:"from,omitempty"`
	To      string `yaml:"to,omitempty"`
	Cc      string `yaml:"cc,omitempty"`
	Subject string `yaml:"subject,omitempty"`
	List    string `yaml:"list,omitempty"`
	Has     string `yaml:"has,omitempty"`
	Query   string `yaml:"query,omitempty"`
}

func (f Filter) Valid(filterNames map[string]struct{}) error {
	// TODO
	return nil
}

type Rule struct {
	Filter  FilterRef `yaml:"filter"`
	Actions Actions   `yaml:"actions"`
}

func (r Rule) Valid(filterNames map[string]struct{}) error {
	// TODO
	return nil
}

type FilterRef struct {
	Name      string `yaml:"name,omitempty"`
	InlineDef Filter `yaml:",inline"`
}

type Actions = v1.Actions
