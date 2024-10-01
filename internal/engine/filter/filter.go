package filter

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/mbrt/gmailctl/internal/engine/gmail"
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

	return w.String()
}

func (fs Filters) DebugString() string {
	w := writer{}

	first := true
	for _, f := range fs {
		if !first {
			w.WriteRune('\n')
		}
		first = false
		w.WriteString(f.DebugString())
	}

	return w.String()
}

// HasLabel returns true if the given label is used by at least one filter.
func (fs Filters) HasLabel(name string) bool {
	for _, f := range fs {
		if f.HasLabel(name) {
			return true
		}
	}
	return false
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

	w.WriteParam("query", indent(f.Criteria.Query, 2))

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
	w.WriteParam("forward to", f.Action.Forward)

	return w.String()
}

// DebugString returns text representation of the filter with extra debugging
// information Gmail search representation and URL included.
func (f Filter) DebugString() string {
	w := writer{}

	w.WriteString(fmt.Sprintf("# Search: %s\n", f.Criteria.ToGmailSearch()))
	w.WriteString(fmt.Sprintf("# URL: %s\n", f.Criteria.ToGmailSearchURL()))
	w.WriteString(f.String())

	return w.String()
}

func indent(query string, level int) string {
	var indented bytes.Buffer
	if !indentInternal(strings.NewReader(query), &indented, level+1) {
		return query
	}
	return "\n" + strings.TrimRight(indented.String(), "\n ")
}

//revive:disable:cyclomatic High complexity, needs some refactoring
func indentInternal(queryReader io.RuneReader, out *bytes.Buffer, level int) bool {
	type parseState int
	const (
		other parseState = iota
		skipSpaces
		inQuotes
	)

	for i := 0; i < level; i++ {
		out.Write([]byte("  "))
	}

	indentationWasNeeded := false
	writeIndentation := func(n int) {
		out.WriteByte('\n')
		for i := 0; i < n; i++ {
			out.Write([]byte("  "))
		}
		indentationWasNeeded = true
	}

	state := skipSpaces

	for {
		r, _, err := queryReader.ReadRune()
		if err != nil {
			break
		}
		switch state {
		case inQuotes:
			out.WriteRune(r)
			if r == '"' {
				state = other
			}
		case skipSpaces, other:
			switch r {
			case ' ':
				if state == skipSpaces {
					continue
				}
				writeIndentation(level)

			case '{', '(':
				out.WriteRune(r)

				level++
				writeIndentation(level)

				state = skipSpaces

			case '}', ')':
				writeIndentation(level - 1)
				out.WriteRune(r)

				level--
				writeIndentation(level)

				state = skipSpaces

			case ':':
				out.WriteByte(':')
				state = skipSpaces

			case '"':
				state = inQuotes
				out.WriteByte('"')

			default:
				if state != inQuotes {
					state = other
				}
				out.WriteRune(r)
			}
		}
	}

	return indentationWasNeeded
}

// HasLabel returns true if the given label is used by the filter.
func (f Filter) HasLabel(name string) bool {
	return f.Action.AddLabel == name
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
	Forward          string
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

// ToGmailSearch returns the equivalent query in Gmail search syntax.
func (c Criteria) ToGmailSearch() string {
	var res []string

	if c.From != "" {
		res = append(res, fmt.Sprintf("from:%s", c.From))
	}
	if c.To != "" {
		res = append(res, fmt.Sprintf("to:%s", c.To))
	}
	if c.Subject != "" {
		res = append(res, fmt.Sprintf("subject:%s", c.Subject))
	}
	if c.Query != "" {
		res = append(res, c.Query)
	}

	return strings.Join(res, " ")
}

// ToGmailSearchURL returns the equivalent query in an URL to Gmail search.
func (c Criteria) ToGmailSearchURL() string {
	return fmt.Sprintf(
		"https://mail.google.com/mail/u/0/#search/%s",
		url.QueryEscape(c.ToGmailSearch()))
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

func (w *writer) String() string {
	return w.b.String()
}

func (w *writer) Err() error {
	return w.err
}
