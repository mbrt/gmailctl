package v1alpha3

import (
	"reflect"
	"strings"

	"github.com/mbrt/gmailctl/internal/engine/gmail"
)

// Version is the latest supported version.
const Version = "v1alpha3"

// Config contains the Jsonnet configuration of the Gmail filters.
type Config struct {
	Version string  `json:"version"`
	Author  Author  `json:"author,omitempty"`
	Labels  []Label `json:"labels,omitempty"`
	Rules   []Rule  `json:"rules"`
	Tests   []Test  `json:"tests,omitempty"`
}

// FilterNode represents a piece of a Gmail filter.
//
// The definition is recursive, as filters can be composed together
// with the use of logical operators. For every filter node, only one
// operator can be specified. If you need to combine multiple queries
// together, combine the nodes with 'And', 'Or' and 'Not'.
type FilterNode struct {
	And []FilterNode `json:"and,omitempty"`
	Or  []FilterNode `json:"or,omitempty"`
	Not *FilterNode  `json:"not,omitempty"`

	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
	Cc      string `json:"cc,omitempty"`
	Bcc     string `json:"bcc,omitempty"`
	ReplyTo string `json:"replyto,omitempty"`
	Subject string `json:"subject,omitempty"`
	List    string `json:"list,omitempty"`
	Has     string `json:"has,omitempty"`
	Query   string `json:"query,omitempty"`

	// IsEscaped specifies that the given parameters don't need any
	// further escaping.
	//
	// Only allowed in combination with 'From', 'To' or 'Subject'.
	IsEscaped bool `json:"isEscaped,omitempty"`
}

// NonEmptyFields returns the names of the fields with a value.
func (f FilterNode) NonEmptyFields() []string {
	// Use reflection to minimize maintenance work.
	var res []string

	v := reflect.ValueOf(f)
	t := reflect.TypeOf(f)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		name := jsonTagName(t.Field(i).Tag)

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
			// Ignore the 'IsEscaped' marker
			continue
		}

		res = append(res, name)
	}

	return res
}

// Rule is the actual complete Gmail filter.
//
// For every email, if the filter applies correctly, then the specified actions
// will be applied to it.
type Rule struct {
	Filter  FilterNode `json:"filter"`
	Actions Actions    `json:"actions"`
}

// Author represents the owner of the gmail account.
type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Actions contains the actions to be applied to a set of emails.
type Actions struct {
	Archive  bool `json:"archive,omitempty"`
	Delete   bool `json:"delete,omitempty"`
	MarkRead bool `json:"markRead,omitempty"`
	Star     bool `json:"star,omitempty"`

	// MarkSpam can be used to disallow mails to be marked as spam.
	// This however is not allowed to be set to true by Gmail.
	MarkSpam      *bool `json:"markSpam,omitempty"`
	MarkImportant *bool `json:"markImportant,omitempty"`

	Category gmail.Category `json:"category,omitempty"`
	Labels   []string       `json:"labels,omitempty"`

	// Forward actions
	Forward string `json:"forward,omitempty"`
}

// Empty returns true if no actions are specified.
func (a Actions) Empty() bool {
	return reflect.DeepEqual(a, Actions{})
}

// Label represents a Gmail label.
type Label struct {
	Name  string      `json:"name"`
	Color *LabelColor `json:"color,omitempty"`
}

// LabelColor is the color of a label.
//
// See https://developers.google.com/gmail/api/v1/reference/users/labels
// for the list of possible colors.
type LabelColor struct {
	Background string `json:"background"`
	Text       string `json:"text"`
}

// Test represents the intended actions applied to a set of emails.
type Test struct {
	// Name is an optional name used for error reporting.
	Name     string    `json:"name,omitempty"`
	Messages []Message `json:"messages"`
	Actions  Actions   `json:"actions"`
}

// Message represents the contents and metadata of an email.
type Message struct {
	From    string   `json:"from,omitempty"`
	To      []string `json:"to,omitempty"`
	Cc      []string `json:"cc,omitempty"`
	Bcc     []string `json:"bcc,omitempty"`
	ReplyTo []string `json:"replyto,omitempty"`
	Lists   []string `json:"lists,omitempty"`
	Subject string   `json:"subject,omitempty"`
	Body    string   `json:"body,omitempty"`
}

func jsonTagName(t reflect.StructTag) string {
	return strings.Split(t.Get("json"), ",")[0]
}
