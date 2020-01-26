package cfgtest

import (
	"strings"

	cfg "github.com/mbrt/gmailctl/pkg/config/v1alpha3"
	"github.com/mbrt/gmailctl/pkg/parser"
)

type matchType int

const (
	matchTypeExact = iota
	matchTypeSuffix
	matchTypeContains
)

type RuleEvaluator interface {
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
	op        parser.FunctionType
	expected  string
	matchType matchType
}

func (n funcNode) Match(msg cfg.Message) bool {
	var fields []string

	switch n.op {
	case parser.FunctionFrom:
		fields = []string{msg.From}
	case parser.FunctionTo:
		fields = msg.To
	case parser.FunctionCc:
		fields = msg.Cc
	case parser.FunctionBcc:
		fields = msg.Bcc
	case parser.FunctionList:
		fields = msg.Lists
	case parser.FunctionSubject:
		fields = []string{msg.Subject}
	case parser.FunctionHas:
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

func normalizeField(a string) string {
	return strings.ToLower(strings.ReplaceAll(a, "@", "."))
}
