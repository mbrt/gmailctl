package v1alpha2

import (
	"reflect"
	"strings"

	v1 "github.com/mbrt/gmailctl/pkg/config/v1alpha1"
	"github.com/mbrt/gmailctl/pkg/gmail"
)

// Version is the latest supported version.
const Version = "v1alpha2"

// Config contains the yaml configuration of the Gmail filters.
type Config struct {
	Version string        `yaml:"version" json:"version"`
	Author  Author        `yaml:"author,omitempty" json:"author,omitempty"`
	Filters []NamedFilter `yaml:"filters,omitempty" json:"filters,omitempty"`
	Rules   []Rule        `yaml:"rules" json:"rules"`
}

// NamedFilter represents a filter with a name.
//
// A named filter can be referenced by other named filters and by filters
// inside rules.
type NamedFilter struct {
	Name  string     `yaml:"name" json:"name"`
	Query FilterNode `yaml:"query" json:"query"`
}

// FilterNode represents a piece of a Gmail filter.
//
// The definition is recursive, as filters can be composed together
// with the use of logical operators. For every filter node, only one
// operator can be specified. If you need to combine multiple queries
// together, combine the nodes with 'And', 'Or' and 'Not'.
type FilterNode struct {
	RefName string `yaml:"name,omitempty" json:"name,omitempty"`

	And []FilterNode `yaml:"and,omitempty" json:"and,omitempty"`
	Or  []FilterNode `yaml:"or,omitempty" json:"or,omitempty"`
	Not *FilterNode  `yaml:"not,omitempty" json:"not,omitempty"`

	From    string `yaml:"from,omitempty" json:"from,omitempty"`
	To      string `yaml:"to,omitempty" json:"to,omitempty"`
	Cc      string `yaml:"cc,omitempty" json:"cc,omitempty"`
	Subject string `yaml:"subject,omitempty" json:"subject,omitempty"`
	List    string `yaml:"list,omitempty" json:"list,omitempty"`
	Has     string `yaml:"has,omitempty" json:"has,omitempty"`
	Query   string `yaml:"query,omitempty" json:"query,omitempty"`

	// IsRaw specifies that no escaping should be done to the given
	// parameters.
	//
	// Only allowed in combination with 'From', 'To' or 'Subject'.
	IsRaw bool `yaml:"isRaw,omitempty" json:"isRaw,omitempty"`
}

// NonEmptyFields returns the names of the fields with a value.
func (f FilterNode) NonEmptyFields() []string {
	// Use reflection to minimize maintenance work.
	var res []string

	v := reflect.ValueOf(f)
	t := reflect.TypeOf(f)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		name := yamlTagName(t.Field(i).Tag)

		switch field.Kind() {
		case reflect.String:
			if field.String() == "" {
				continue
			}
		case reflect.Slice:
			if field.Len() == 0 {
				continue
			}
		case reflect.Ptr:
			if field.Pointer() == 0 {
				continue
			}
		case reflect.Bool:
			// Ignore the 'Raw' marker
			continue
		}

		res = append(res, name)
	}

	return res
}

// Empty returns true if all the fields are empty.
func (f FilterNode) Empty() bool {
	// Use reflection to minimize maintenance work.
	count := 0

	v := reflect.ValueOf(f)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		switch field.Kind() {
		case reflect.String:
			if field.String() == "" {
				continue
			}
		case reflect.Slice:
			if field.Len() == 0 {
				continue
			}
		case reflect.Ptr:
			if field.Pointer() == 0 {
				continue
			}
		case reflect.Bool:
			// Ignore the 'Raw' marker
			continue
		}

		count++
	}

	return count == 0
}

// Rule is the actual complete Gmail filter.
//
// For every email, if the filter applies correctly, then the specified actions
// will be applied to it.
type Rule struct {
	Filter  FilterNode `yaml:"filter" json:"filter"`
	Actions Actions    `yaml:"actions" json:"actions"`
}

// Author represents the owner of the gmail account.
type Author v1.Author

// Actions contains the actions to be applied to a set of emails.
type Actions struct {
	Archive  bool `yaml:"archive,omitempty" json:"archive,omitempty"`
	Delete   bool `yaml:"delete,omitempty" json:"delete,omitempty"`
	MarkRead bool `yaml:"markRead,omitempty" json:"markRead,omitempty"`
	Star     bool `yaml:"star,omitempty" json:"star,omitempty"`

	// MarkSpam can be used to disallow mails to be marked as spam.
	// This however is not allowed to be set to true by Gmail.
	MarkSpam      *bool `yaml:"markSpam,omitempty" json:"markSpam,omitempty"`
	MarkImportant *bool `yaml:"markImportant,omitempty" json:"markImportant,omitempty"`

	Category gmail.Category `yaml:"category,omitempty" json:"category,omitempty"`
	Labels   []string       `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// Empty returns true if no actions are specified.
func (a Actions) Empty() bool {
	return reflect.DeepEqual(a, Actions{})
}

func yamlTagName(t reflect.StructTag) string {
	return strings.Split(t.Get("yaml"), ",")[0]
}
