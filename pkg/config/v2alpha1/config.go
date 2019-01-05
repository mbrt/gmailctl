package v2alpha1

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"

	v1 "github.com/mbrt/gmailctl/pkg/config/v1alpha1"
)

// Version is the latest supported version.
const Version = "v2alpha1"

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

	var filters NamesSet
	for _, f := range c.Filters {
		if err := f.Valid(filters); err != nil {
			return errors.Wrap(err, "invalid filter")
		}
		filters[f.Name] = struct{}{}
	}

	for _, r := range c.Rules {
		if err := r.Valid(filters); err != nil {
			return errors.Wrap(err, "invalid rule")
		}
	}

	return nil
}

// NamedFilter represents a filter with a name.
//
// A named filter can be referenced by other named filters and by filters
// inside rules.
type NamedFilter struct {
	Name  string     `yaml:"name"`
	Query FilterNode `yaml:"query"`
}

// Valid returns an error if the configuration is invalid.
func (f NamedFilter) Valid(otherFilters NamesSet) error {
	if f.Name == "" {
		return errors.New("invalid empty filter name")
	}
	return f.Query.Valid(otherFilters)
}

// FilterNode represents a piece of a Gmail filter.
//
// The definition is recursive, as filters can be composed together
// with the use of logical operators. For every filter node, only
// operator can be specified. If you need to combine multiple queries
// together, combine the nodes with 'And', 'Or' and 'Not'.
type FilterNode struct {
	RefName string `yaml:"name,omitempty"`

	And []FilterNode `yaml:"and,omitempty"`
	Or  []FilterNode `yaml:"or,omitempty"`
	Not *FilterNode  `yaml:"not,omitempty"`

	From    string `yaml:"from,omitempty"`
	To      string `yaml:"to,omitempty"`
	Cc      string `yaml:"cc,omitempty"`
	Subject string `yaml:"subject,omitempty"`
	List    string `yaml:"list,omitempty"`
	Has     string `yaml:"has,omitempty"`
	Query   string `yaml:"query,omitempty"`
}

// Valid returns an error if the configuration is invalid.
func (f FilterNode) Valid(filters NamesSet) error {
	// Use reflection to minimize maintenance work.
	var nonEmpty []string

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
			if field.Len() == 1 {
				return errors.Errorf("%s: only one filter specified for binary operator", name)
			}
			for i = 0; i < field.Len(); i++ {
				subfilter := field.Index(i).Interface().(FilterNode)
				if err := subfilter.Valid(filters); err != nil {
					return errors.Wrapf(err, "inside '%s'", name)
				}
			}
		case reflect.Ptr:
			if field.Pointer() == 0 {
				continue
			}
			if err := field.Interface().(*FilterNode).Valid(filters); err != nil {
				return errors.Wrapf(err, "inside '%s'", name)
			}
		}

		nonEmpty = append(nonEmpty, name)
	}

	// Check RefName explicitly
	if _, ok := filters[f.RefName]; f.RefName != "" && !ok {
		return errors.Errorf("invalid filter reference '%s', not found", f.RefName)
	}

	if len(nonEmpty) > 1 {
		return errors.Errorf("invalid multiple fields specified without combining operator (and/or/not): %s",
			strings.Join(nonEmpty, ","))
	}

	if len(nonEmpty) == 0 {
		return errors.New("empty filter")
	}

	return nil
}

// Empty returns true if the filter does not contain any action.
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
	Filter  FilterNode `yaml:"filter"`
	Actions Actions    `yaml:"actions"`
}

// Valid returns an error if the configuration is invalid.
func (r Rule) Valid(filters NamesSet) error {
	if err := r.Filter.Valid(filters); err != nil {
		return errors.Wrap(err, "invalid filter in rule")
	}
	if r.Actions.Empty() {
		return errors.New("no actions in rule")
	}
	return nil
}

// Author represents the owner of the gmail account.
type Author v1.Author

// Actions contains the actions to be applied to a set of emails.
type Actions v1.Actions

// Empty returns true if no actions are specified.
func (a Actions) Empty() bool {
	return reflect.DeepEqual(a, Actions{})
}

// NamesSet is a set of names
type NamesSet map[string]struct{}

func yamlTagName(t reflect.StructTag) string {
	return strings.Split(t.Get("yaml"), ",")[0]
}
