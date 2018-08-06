package main

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

const (
	PropertyFrom          = "from"
	PropertyTo            = "to"
	PropertyHas           = "hasTheWord"
	PropertyDoesNotHave   = "doesNotHaveTheWord"
	PropertyMarkImportant = "shouldAlwaysMarkAsImportant"
	PropertyApplyLabel    = "label"
	PropertyApplyCategory = "smartLabelToApply"
	PropertyDelete        = "shouldTrash"
)

type Entry struct {
	Properties []Property
}

type Property struct {
	Name  string
	Value string
}

func GenerateRules(config Config) ([]Entry, error) {
	res := []Entry{}
	for i, rule := range config.Rules {
		entries, err := generateRule(rule, config.Consts)
		if err != nil {
			return res, errors.Wrap(err, fmt.Sprintf("error generating rule #%d", i))
		}
		res = append(res, entries...)
	}
	return res, nil
}

func generateRule(rule Rule, consts Consts) ([]Entry, error) {
	filters, err := generateFilters(rule.Filters, consts)
	if err != nil {
		return nil, errors.Wrap(err, "error generating filters")
	}
	if len(filters) == 0 {
		return nil, errors.New("at least one filter has to be specified")
	}
	actions, err := generateActions(rule.Actions, consts)
	if err != nil {
		return nil, errors.Wrap(err, "error generating actions")
	}
	if len(actions) == 0 {
		return nil, errors.New("at least one action has to be specified")
	}
	return combineFiltersActions(filters, actions), nil
}

func generateFilters(filters Filters, consts Consts) ([]Property, error) {
	res := []Property{}
	if len(filters.From) > 0 {
		p := Property{PropertyFrom, joinOR(filters.From)}
		res = append(res, p)
	}
	if len(filters.Subject) > 0 {
		p := Property{PropertyFrom, joinOR(filters.Subject)}
		res = append(res, p)
	}
	if len(filters.To) > 0 {
		p := Property{PropertyTo, joinOR(filters.To)}
		res = append(res, p)
	}
	// TODO To, NotTo
	// The negation looks like:
	// -{to:{foobar@baz.com} } -{"Build failed"}
	// which are mapped to hasTheWord and doesNotHaveTheWord
	return res, nil
}

func generateActions(actions Actions, consts Consts) ([]Property, error) {
	return nil, nil
}

func joinOR(a []string) string {
	if containsSpace(a) {
		a = quote(a)
	}
	return fmt.Sprintf("{%s}", strings.Join(a, " "))
}

func containsSpace(a []string) bool {
	for _, s := range a {
		if strings.ContainsRune(s, ' ') {
			return true
		}
	}
	return false
}

func quote(a []string) []string {
	res := make([]string, len(a))
	for i, s := range a {
		res[i] = fmt.Sprintf(`"%s"`, s)
	}
	return res
}

func combineFiltersActions(filters []Property, actions []Property) []Entry {
	return nil
}
