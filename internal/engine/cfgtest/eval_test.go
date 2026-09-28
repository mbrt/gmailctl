package cfgtest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cfg "github.com/mbrt/gmailctl/internal/engine/config/v1alpha3"
	"github.com/mbrt/gmailctl/internal/engine/parser"
)

func TestEmailFieldMatching(t *testing.T) {
	fields := []struct {
		function parser.FunctionType
		message  func(string) cfg.Message
	}{
		{parser.FunctionFrom, func(s string) cfg.Message { return cfg.Message{From: s} }},
		{parser.FunctionTo, func(s string) cfg.Message { return cfg.Message{To: []string{s}} }},
		{parser.FunctionCc, func(s string) cfg.Message { return cfg.Message{Cc: []string{s}} }},
		{parser.FunctionBcc, func(s string) cfg.Message { return cfg.Message{Bcc: []string{s}} }},
		{parser.FunctionReplyTo, func(s string) cfg.Message { return cfg.Message{ReplyTo: []string{s}} }},
		{parser.FunctionDeliveredTo, func(s string) cfg.Message { return cfg.Message{DeliveredTo: []string{s}} }},
		{parser.FunctionList, func(s string) cfg.Message { return cfg.Message{Lists: []string{s}} }},
		{parser.FunctionHas, func(s string) cfg.Message { return cfg.Message{From: s} }},
	}
	tests := []struct {
		name    string
		filter  string
		address string
		match   bool
	}{
		{"bare name", "some-list", "some-list", true},
		{"local part", "some-list", "some-list@google.com", true},
		{"case insensitive", "SOME-LIST", "Some-List@Google.com", true},
		{"domain", "google.com", "some-list@google.com", true},
		{"dotted name", "some", "some.list@google.com", true},
		{"full address", "me@gmail.com", "me@gmail.com", true},
		{"at and dot equivalent", "some-list.google.com", "some-list@google.com", true},
		{"domain suffix", "@google.com", "some-list@google.com", true},
		{"wildcard domain", "*@google.com", "some-list@google.com", true},
		{"plus address", "some+list@google.com", "some+list@google.com", true},
		{"quoted address", `"some+list@google.com"`, "some+list@google.com", true},
		{"unicode name", "büro", "büro@example.com", true},
		{"later whole word", "me", "notme.me@gmail.com", true},
		{"prefix of name", "some", "someone@google.com", false},
		{"suffix of name", "me", "notme@gmail.com", false},
		{"suffix of full address", "me@gmail.com", "notme@gmail.com", false},
		{"number is part of name", "list", "list2@google.com", false},
		{"underscore is part of name", "list", "list_name@google.com", false},
		{"unicode is part of name", "me", "éme@example.com", false},
		{"combining mark is part of name", "me", "me\u0301@example.com", false},
		{"literal dot", "some.list", "someXlist@google.com", false},
		{"literal plus", "some+list@google.com", "somelist@google.com", false},
		{"different domain", "@google.com", "some-list@notgoogle.com", false},
		{"domain must be suffix", "@google.com", "some-list@google.com.example.org", false},
		{"empty field", "some-list", "", false},
	}

	for _, field := range fields {
		t.Run(field.function.String(), func(t *testing.T) {
			for _, tc := range tests {
				t.Run(tc.name, func(t *testing.T) {
					eval, err := NewEvaluator(fn1(field.function, tc.filter))
					require.Nil(t, err)
					require.Equal(t, tc.match, eval.Match(field.message(tc.address)))
				})
			}
		})
	}
}

func TestParseEval(t *testing.T) {
	expr := or(
		and(
			fn(parser.FunctionList, parser.OperationOr,
				"list1.gm.com", "list2@gm.com"),
			fn1(parser.FunctionTo, "me@gmail.com"),
		),
		fn1(parser.FunctionSubject, "Subject"),
		fn1(parser.FunctionSubject, `"Exact phrase"`),
		fn(parser.FunctionFrom, parser.OperationOr, "@google.com", "b"),
		fn1(parser.FunctionHas, "Important message"),
		fn1(parser.FunctionHas, "foo@bar.com"),
	)
	eval, err := NewEvaluator(expr)
	if err != nil {
		t.Fatalf("NewEvaluator failed: %v", err)
	}

	tests := []struct {
		name        string
		message     cfg.Message
		expectMatch bool
	}{
		{
			name: "subject",
			message: cfg.Message{
				Subject: "contains subject yes",
			},
			expectMatch: true,
		},
		{
			name: "quoted subject",
			message: cfg.Message{
				Subject: "has an exact phrase inside",
			},
			expectMatch: true,
		},
		{
			name: "list with @",
			message: cfg.Message{
				Lists: []string{"list1@gm.com"},
				To:    []string{"me@gmail.com"},
			},
			expectMatch: true,
		},
		{
			name: "from google",
			message: cfg.Message{
				From: "someone@google.com",
			},
			expectMatch: true,
		},
		{
			name: "has from",
			message: cfg.Message{
				From: "foo@bar.com",
			},
			expectMatch: true,
		},
		{
			name: "has body",
			message: cfg.Message{
				Body: "important message",
			},
			expectMatch: true,
		},
		{
			name: "list but not to me",
			message: cfg.Message{
				Lists: []string{"list1@gm.com"},
				To:    []string{"notme@gmail.com"},
			},
			expectMatch: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			match := eval.Match(tc.message)
			assert.Equal(t, tc.expectMatch, match)
		})
	}
}
