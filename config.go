package main

// Config contains the yaml configuration of the Gmail filters.
type Config struct {
	Version string            `yaml:"version"`
	Author  Author            `yaml:"author"`
	Consts  map[string]Values `yaml:"consts,omitempty"`
	Rules   []Rule            `yaml:"rules"`
}

// Author represents the owner of the gmail account.
type Author struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

// Values is a container for an array of string values.
type Values struct {
	Values []string `yaml:"values"`
}

// Rule is a filter with an associated action.
type Rule struct {
	Filters Filters `yaml:"filters"`
	Actions Actions `yaml:"actions"`
}

// Filters determine how to select emails.
//
// Two ways are possible: by directly specifying the constants or by referring
// to constants specified by the `consts` section of the config.
type Filters struct {
	SimpleFilters `yaml:",inline"`
	Consts        SimpleFilters `yaml:"consts"`
}

// SimpleFilters contains a list of filters interpreted at a higher level.
//
// Every type of filter (e.g. Subject) is an array or requirements. They will be OR-ed
// together. If multiple types of filters are specified, they will be put in AND together.
type SimpleFilters struct {
	Subject []string `yaml:"subject,omitempty"`
	From    []string `yaml:"from,omitempty"`
	To      []string `yaml:"to,omitempty"`
	NotTo   []string `yaml:"notTo,omitempty"`
}

// Actions contains the actions to be applied to a set of emails.
type Actions struct {
	Archive       bool     `yaml:"archive,omitempty"`
	Delete        bool     `yaml:"delete,omitempty"`
	MarkImportant bool     `yaml:"markImportant,omitempty"`
	MarkRead      bool     `yaml:"markRead,omitempty"`
	Labels        []string `yaml:"labels,omitempty"`
}
