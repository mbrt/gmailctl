package filter

import "github.com/mbrt/gmailfilter/pkg/config"

// Filters is a list of filters created in Gmail.
type Filters []Filter

// Filter matches 1:1 a filter created on Gmail.
type Filter struct {
	// ID is an optional identifier associated with a filter.
	ID       string
	Action   Action
	Criteria Criteria
}

// Action represents an action associated with a Gmail filter.
type Action struct {
	Archive       bool
	Delete        bool
	MarkImportant bool
	MarkRead      bool
	Category      config.Category
	AddLabel      string
}

// Empty returns true if no action is specified.
func (a Action) Empty() bool {
	return a == Action{}
}

// Criteria represents the filtering criteria associated with a Gmail filter.
type Criteria struct {
	From    string
	To      string
	Subject string
	Query   string
}

// Empty returns true if no criteria is specified.
func (c Criteria) Empty() bool {
	return c == Criteria{}
}

// Label contains information about a Gmail label.
type Label struct {
	ID   string
	Name string
}
