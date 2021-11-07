package v1alpha1

import (
	"github.com/mbrt/gmailctl/internal/engine/gmail"
)

// Version is the config format version supported by this package.
const Version = "v1alpha1"

// Config contains the yaml configuration of the Gmail filters.
type Config struct {
	Version string `yaml:"version"`
	Author  Author `yaml:"author"`
	Consts  Consts `yaml:"consts,omitempty"`
	Rules   []Rule `yaml:"rules"`
}

// Consts maps names to a list of string values
type Consts map[string]ConstValue

// Author represents the owner of the gmail account.
type Author struct {
	Name  string `yaml:"name" json:"name"`
	Email string `yaml:"email" json:"email"`
}

// ConstValue is a container for an array of string values.
type ConstValue struct {
	Values []string `yaml:"values"`
}

// Rule is a filter with an associated action.
type Rule struct {
	Filters Filters `yaml:"filters"`
	Actions Actions `yaml:"actions"`
}

// Filters determine how to select emails.
//
// Two ways are possible: by directly specifying the constants in the match filters
// or by referring to constants specified by the global `consts` section on top of the
// config.
type Filters struct {
	CompositeFilters `yaml:",inline"`
	Consts           CompositeFilters `yaml:"consts"`
	Query            string           `yaml:"query"`
}

// CompositeFilters contains alternatively match or negation of matches.
//
// All the conditions are put in AND together.
type CompositeFilters struct {
	MatchFilters `yaml:",inline"`
	Not          MatchFilters `yaml:"not"`
}

// MatchFilters contains a list of filters interpreted at a higher level.
//
// Every type of filter (e.g. Subject) is an array or requirements. They will be OR-ed
// together. If multiple types of filters are specified, they will be put in AND together.
type MatchFilters struct {
	From    []string `yaml:"from,omitempty"`
	To      []string `yaml:"to,omitempty"`
	Cc      []string `yaml:"cc,omitempty"`
	Subject []string `yaml:"subject,omitempty"`
	Has     []string `yaml:"has,omitempty"`
	List    []string `yaml:"list,omitempty"`
}

// Actions contains the actions to be applied to a set of emails.
type Actions struct {
	Archive       bool           `yaml:"archive,omitempty"`
	Delete        bool           `yaml:"delete,omitempty"`
	MarkImportant bool           `yaml:"markImportant,omitempty"`
	MarkRead      bool           `yaml:"markRead,omitempty"`
	Category      gmail.Category `yaml:"category,omitempty"`
	Labels        []string       `yaml:"labels,omitempty"`
}
