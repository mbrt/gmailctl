package filter

import (
	"strings"

	"github.com/mbrt/gmailctl/pkg/gmail"
)

// Filters is a list of filters created in Gmail.
type Filters []Filter

func (fs Filters) String() string {
	w := writer{}

	first := true
	for _, f := range fs {
		if !first {
			w.WriteRune('\n')
		}
		first = false
		w.WriteString(f.String())
	}

	str, _ := w.Result()
	return str
}

// Filter matches 1:1 a filter created on Gmail.
type Filter struct {
	// ID is an optional identifier associated with a filter.
	ID       string
	Action   Actions
	Criteria Criteria
}

func (f Filter) String() string {
	w := writer{}

	w.WriteString("* Criteria:\n")
	w.WriteParam("from", f.Criteria.From)
	w.WriteParam("to", f.Criteria.To)
	w.WriteParam("subject", f.Criteria.Subject)
	w.WriteParam("query", f.Criteria.Query)

	w.WriteString("  Actions:\n")
	w.WriteBool("archive", f.Action.Archive)
	w.WriteBool("delete", f.Action.Delete)
	w.WriteBool("mark as important", f.Action.MarkImportant)
	w.WriteBool("never mark as important", f.Action.MarkNotImportant)
	w.WriteBool("never mark as spam", f.Action.MarkNotSpam)
	w.WriteBool("mark as read", f.Action.MarkRead)
	w.WriteBool("star", f.Action.Star)
	w.WriteParam("categorize as", string(f.Action.Category))
	w.WriteParam("apply label", f.Action.AddLabel)

	str, _ := w.Result()
	return str
}

// Actions represents an action associated with a Gmail filter.
type Actions struct {
	AddLabel         string
	Category         gmail.Category
	Archive          bool
	Delete           bool
	MarkImportant    bool
	MarkNotImportant bool
	MarkRead         bool
	MarkNotSpam      bool
	Star             bool
}

// Empty returns true if no action is specified.
func (a Actions) Empty() bool {
	return a == Actions{}
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

type writer struct {
	b   strings.Builder
	err error
}

func (w *writer) WriteParam(name, value string) {
	if value == "" {
		return
	}
	w.WriteString("    ")
	w.WriteString(name)
	w.WriteString(": ")
	w.WriteString(value)
	w.WriteRune('\n')
}

func (w *writer) WriteBool(name string, value bool) {
	if !value {
		return
	}
	w.WriteString("    ")
	w.WriteString(name)
	w.WriteRune('\n')
}

func (w *writer) WriteString(a string) {
	if w.err != nil {
		return
	}
	_, w.err = w.b.WriteString(a)
}

func (w *writer) WriteRune(a rune) {
	if w.err != nil {
		return
	}
	_, w.err = w.b.WriteRune(a)
}

func (w *writer) Result() (string, error) {
	return w.b.String(), w.err
}
