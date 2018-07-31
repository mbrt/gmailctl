package main

type Config struct {
	Version string            `yaml:"version"`
	Consts  map[string]Values `yaml:"consts,omitempty"`
	Rules   []Rule            `yaml:"rules"`
}

type Values struct {
	Values []string `yaml:"values"`
}

type Rule struct {
	Filters Filters `yaml:"filters"`
	Actions Actions `yaml:"actions"`
}

type Filters struct {
	SimpleFilters `yaml:",inline"`
	Consts        SimpleFilters `yaml:"consts"`
}

type SimpleFilters struct {
	Subject []string `yaml:"subject,omitempty"`
	From    []string `yaml:"from,omitempty"`
	To      []string `yaml:"to,omitempty"`
	NotTo   []string `yaml:"notTo,omitempty"`
}

type Actions struct {
	Archive       bool     `yaml:"archive,omitempty"`
	Delete        bool     `yaml:"delete,omitempty"`
	MarkImportant bool     `yaml:"markImportant,omitempty"`
	MarkRead      bool     `yaml:"markRead,omitempty"`
	Labels        []string `yaml:"labels,omitempty"`
}
