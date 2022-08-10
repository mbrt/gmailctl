package cfgtest

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/pmezard/go-difflib/difflib"

	"github.com/mbrt/gmailctl/internal/engine/config/v1alpha3"
	"github.com/mbrt/gmailctl/internal/engine/gmail"
	"github.com/mbrt/gmailctl/internal/engine/parser"
	"github.com/mbrt/gmailctl/internal/errors"
	"github.com/mbrt/gmailctl/internal/reporting"
)

// NewFromParserRules translates parser Rules into test Rules.
//
// This function is best effort. Every criteria that is not convertible is going
// to be ignored and an error is returned in its place. The resulting rules will
// contain only the valid rules.
func NewFromParserRules(rs []parser.Rule) (Rules, error) {
	var res Rules
	var errs error

	for i, pr := range rs {
		re, err := NewEvaluator(pr.Criteria)
		if err != nil {
			errs = errors.Combine(
				errs,
				fmt.Errorf("cannot evaluate criteria #%d: %w", i, err),
			)
			continue
		}
		res = append(res, Rule{re, Actions(pr.Actions)})
	}

	return res, errs
}

// Rule represents a filter that can evaluate whether messages apply to it.
type Rule struct {
	Eval    RuleEvaluator
	Actions Actions
}

// Rules is a set of rules.
type Rules []Rule

// ExecTests evaluates all the rules against the given tests.
//
// The evaluation stops at the first failing test.
func (rs Rules) ExecTests(ts []v1alpha3.Test) Result {
	var failed []FailedTest

	for i, t := range ts {
		if errs := rs.ExecTest(t); len(errs) > 0 {
			failed = append(failed, FailedTest{
				ID:     i,
				Name:   t.Name,
				Errors: errs,
			})
		}
	}

	return Result{
		OK:       len(failed) == 0,
		NumTests: len(ts),
		Failed:   failed,
	}
}

// ExecTest evaluates the rules on all the messages of the given test.
//
// If the rules apply as expected by the test, no error is returned.
func (rs Rules) ExecTest(t v1alpha3.Test) []error {
	var res error

	for i, msg := range t.Messages {
		expected, err := rs.MatchingActions(msg)
		if err != nil {
			res = errors.Combine(
				res,
				errors.WithDetails(
					fmt.Errorf("message #%d: error evaluating matching filters: %w", i, err),
					messageDetails(msg)),
			)
			continue
		}
		if expected.Equal(Actions(t.Actions)) {
			// All good with this message.
			continue
		}

		// Report the error.
		diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(reporting.Prettify(t.Actions, false)),
			B:        difflib.SplitLines(reporting.Prettify(expected, false)),
			FromFile: "want",
			ToFile:   "got",
			Context:  5,
		})
		if err != nil {
			// The diff failing is not a big deal, but we should return
			// something.
			diff = fmt.Sprintf("<cannot compute diff>: %v", err)
		}
		res = errors.Combine(
			res,
			errors.WithDetails(
				fmt.Errorf("message #%d is going to get unexpected actions: %s", i,
					reporting.Prettify(expected, true)),
				messageDetails(msg),
				fmt.Sprintf("Actions:\n%s", strings.TrimRight(diff, "\n"))),
		)
	}

	return errors.Errors(res)
}

// MatchingActions returns the actions that would be applied by the rules if
// the given message arrived.
//
// An error can be returned if multiple incompatible actions would be applied. Note
// that in Gmail this wouldn't be an error, but a nondeterministic action would be
// applied. Since this situation is most likely a mistake by the user, we treat it
// as an error.
func (rs Rules) MatchingActions(msg v1alpha3.Message) (Actions, error) {
	var (
		res Actions
		err error
	)
	for _, rule := range rs {
		if rule.Eval.Match(msg) {
			if res, err = mergeActions(res, rule.Actions); err != nil {
				return res, fmt.Errorf("conflicting filters detected: %w", err)
			}
		}
	}
	return res, nil
}

// Result represents the result of a series of tests.
type Result struct {
	OK       bool
	NumTests int
	Failed   []FailedTest
}

func (r Result) String() string {
	var buf bytes.Buffer

	if r.OK {
		fmt.Fprintf(&buf, "Success: %d/%d", r.NumTests, r.NumTests)
		return buf.String()
	}
	fmt.Fprintf(&buf, "Failed: %d/%d\n", len(r.Failed), r.NumTests)
	for _, t := range r.Failed {
		t.dump(&buf)
	}

	return buf.String()
}

// FailedTest includes all the errors of a failed test.
type FailedTest struct {
	ID     int
	Name   string
	Errors []error
}

func (t FailedTest) String() string {
	var buf bytes.Buffer
	t.dump(&buf)
	return buf.String()
}

func (t FailedTest) dump(w io.Writer) {
	name := t.Name
	if name == "" {
		name = fmt.Sprintf("#%d", t.ID)
	}
	fmt.Fprintf(w, "\nFailed test %q:\n%+v\n", name, errors.Combine(t.Errors...))
}

// Actions represent the actions applied by a filter.
type Actions parser.Actions

// Equal returns true if the given actions are equivalent to this object.
func (a Actions) Equal(a2 Actions) bool {
	if a.Archive != a2.Archive {
		return false
	}
	if a.Delete != a2.Delete {
		return false
	}
	if a.MarkRead != a2.MarkRead {
		return false
	}
	if a.Star != a2.Star {
		return false
	}
	if !triboolsEqual(a.MarkSpam, a2.MarkSpam) {
		return false
	}
	if !triboolsEqual(a.MarkImportant, a2.MarkImportant) {
		return false
	}
	if a.Category != a2.Category {
		return false
	}
	if !stringSliceEqual(a.Labels, a2.Labels) {
		return false
	}
	return a.Forward == a2.Forward
}

func mergeActions(a1, a2 Actions) (Actions, error) {
	res := Actions{
		Archive:  a1.Archive || a2.Archive,
		Delete:   a1.Delete || a2.Delete,
		MarkRead: a1.MarkRead || a2.MarkRead,
		Star:     a1.Star || a2.Star,
	}
	var err error
	if res.MarkSpam, err = mergeTribool(a1.MarkSpam, a2.MarkSpam); err != nil {
		return res, fmt.Errorf("'markSpam' is applied differently: %w", err)
	}
	if res.MarkImportant, err = mergeTribool(a1.MarkImportant, a2.MarkImportant); err != nil {
		return res, fmt.Errorf("'markImportant' is applied differently: %w", err)
	}
	if res.Category, err = mergeCategories(a1.Category, a2.Category); err != nil {
		return res, fmt.Errorf("'category' is applied differently: %w", err)
	}
	res.Labels = append(res.Labels, a1.Labels...)
	res.Labels = append(res.Labels, a2.Labels...)
	if res.Forward, err = mergeStrings(a1.Forward, a2.Forward); err != nil {
		return res, fmt.Errorf("'forward' is applied differently: %w", err)
	}
	return res, nil
}

func mergeTribool(b1, b2 *bool) (*bool, error) {
	if b1 == nil {
		return b2, nil
	}
	if b2 == nil {
		return b1, nil
	}
	if *b1 != *b2 {
		return nil, fmt.Errorf("got %v and %v", *b1, *b2)
	}
	return b1, nil
}

func mergeCategories(c1, c2 gmail.Category) (gmail.Category, error) {
	r, err := mergeStrings(string(c1), string(c2))
	return gmail.Category(r), err
}

func mergeStrings(s1, s2 string) (string, error) {
	if s1 == "" {
		return s2, nil
	}
	if s2 == "" {
		return s1, nil
	}
	if s1 != s2 {
		return "", fmt.Errorf("got %s and %s", s1, s2)
	}
	return s1, nil
}

func triboolsEqual(b1, b2 *bool) bool {
	if b1 == nil && b2 == nil {
		return true
	}
	if b1 == nil {
		return b2 == nil
	}
	if b2 == nil {
		return b1 == nil
	}
	return *b1 == *b2
}

func stringSliceEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	sort.Strings(s1)
	sort.Strings(s2)
	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

func messageDetails(msg v1alpha3.Message) string {
	return fmt.Sprintf("Message: %s", reporting.Prettify(msg, false))
}
