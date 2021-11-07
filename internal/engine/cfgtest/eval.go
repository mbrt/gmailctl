package cfgtest

import (
	"strings"

	cfg "github.com/mbrt/gmailctl/internal/engine/config/v1alpha3"
)

type matchType int

const (
	matchTypeExact = iota
	matchTypeSuffix
	matchTypeContains
)

type matchField int

const (
	matchFieldUnknown = iota
	matchFieldFrom
	matchFieldTo
	matchFieldCc
	matchFieldBcc
	matchFieldReplyTo
	matchFieldLists
	matchFieldSubject
	matchFieldBody
)

// RuleEvaluator represents a filter criteria able to evaluate if an email matches
// its definition.
type RuleEvaluator interface {
	// Match returns true if the given message matches the filter criteria.
	Match(msg cfg.Message) bool
}

type andNode struct {
	children []RuleEvaluator
}

func (n andNode) Match(msg cfg.Message) bool {
	for _, c := range n.children {
		if !c.Match(msg) {
			return false
		}
	}
	return true
}

type orNode struct {
	children []RuleEvaluator
}

func (n orNode) Match(msg cfg.Message) bool {
	for _, c := range n.children {
		if c.Match(msg) {
			return true
		}
	}
	return false
}

type notNode struct {
	child RuleEvaluator
}

func (n notNode) Match(msg cfg.Message) bool {
	return !n.child.Match(msg)
}

type funcNode struct {
	field     matchField
	expected  string
	matchType matchType
}

func (n funcNode) Match(msg cfg.Message) bool {
	var fields []string

	switch n.field {
	case matchFieldFrom:
		fields = []string{msg.From}
	case matchFieldTo:
		fields = msg.To
	case matchFieldCc:
		fields = msg.Cc
	case matchFieldBcc:
		fields = msg.Bcc
	case matchFieldReplyTo:
		fields = msg.ReplyTo
	case matchFieldLists:
		fields = msg.Lists
	case matchFieldSubject:
		fields = []string{msg.Subject}
	case matchFieldBody:
		fields = []string{msg.Body}
	}

	for _, f := range fields {
		isMatch := false
		normF := normalizeField(f)

		switch n.matchType {
		case matchTypeExact:
			isMatch = normF == n.expected
		case matchTypeSuffix:
			isMatch = strings.HasSuffix(normF, n.expected)
		case matchTypeContains:
			isMatch = strings.Contains(normF, n.expected)
		}

		// If there's no match, continue searching.
		if isMatch {
			return true
		}
	}

	return false
}

// normalizeField emulates Gmail normalization: @ and . are the same, and
// the match is case insensitive.
func normalizeField(a string) string {
	return strings.ToLower(strings.ReplaceAll(a, "@", "."))
}
