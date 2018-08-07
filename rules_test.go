package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchFilter(t *testing.T) {
	// Test a single filter
	filt := MatchFilters{
		Subject: []string{"important", "not important"},
	}
	props, err := generateMatchFilters(filt)
	assert.Nil(t, err)
	expected := []Property{
		{PropertySubject, `{important "not important"}`},
	}
	assert.Equal(t, expected, props)

	// Test all the filters together
	filt = MatchFilters{
		From:    []string{"foobar@mail.com", "baz@g.com"},
		To:      []string{"my@self.com"},
		Subject: []string{"important", "not important"},
		Has:     []string{"what's wrong", "alert"},
	}
	props, err = generateMatchFilters(filt)
	assert.Nil(t, err)
	expected = []Property{
		{PropertyFrom, "{foobar@mail.com baz@g.com}"},
		{PropertyTo, "my@self.com"},
		{PropertySubject, `{important "not important"}`},
		{PropertyHas, `{"what's wrong" alert}`},
	}
	assert.Equal(t, expected, props)
}

func TestActions(t *testing.T) {
	// Test all the actions together
	act := Actions{
		Archive:       true,
		Delete:        true,
		MarkImportant: true,
		MarkRead:      true,
		Category:      CategoryPersonal,
		Labels:        []string{"label1", "label2"},
	}
	props, err := generateActions(act)
	assert.Nil(t, err)
	expected := []Property{
		{PropertyArchive, "true"},
		{PropertyDelete, "true"},
		{PropertyMarkImportant, "true"},
		{PropertyMarkRead, "true"},
		{PropertyApplyCategory, "^smartlabel_personal"},
		{PropertyApplyLabel, "label1"},
		{PropertyApplyLabel, "label2"},
	}
	assert.Equal(t, expected, props)
}

func TestGenerateSingleEntry(t *testing.T) {
	// Smoke test with a single entry as result
	mf := MatchFilters{
		From: []string{"foobar@mail.com"},
	}
	actions := Actions{
		Archive:  true,
		MarkRead: true,
	}
	config := Config{
		Rules: []Rule{{ /* single empty rule */ }},
	}
	config.Rules[0].Filters.MatchFilters = mf
	config.Rules[0].Actions = actions

	entries, err := GenerateRules(config)
	assert.Nil(t, err)
	expected := []Entry{
		Entry{
			Property{PropertyFrom, "foobar@mail.com"},
			Property{PropertyArchive, "true"},
			Property{PropertyMarkRead, "true"},
		},
	}
	assert.Equal(t, expected, entries)
}

func TestGenerateMultipleEntities(t *testing.T) {
	// Smoke test with a single entry as result
	mf := MatchFilters{
		From: []string{"foobar@mail.com"},
		Has:  []string{"pippo", "pluto paperino"},
	}
	actions := Actions{
		MarkRead: true,
		Labels:   []string{"label1", "label2", "label3"},
	}
	config := Config{
		Rules: []Rule{{ /* single empty rule */ }},
	}
	config.Rules[0].Filters.MatchFilters = mf
	config.Rules[0].Actions = actions

	entries, err := GenerateRules(config)
	assert.Nil(t, err)
	expected := []Entry{
		Entry{
			Property{PropertyFrom, "foobar@mail.com"},
			Property{PropertyHas, `{pippo "pluto paperino"}`},
			Property{PropertyMarkRead, "true"},
			Property{PropertyApplyLabel, "label1"},
		},
		Entry{
			Property{PropertyFrom, "foobar@mail.com"},
			Property{PropertyHas, `{pippo "pluto paperino"}`},
			Property{PropertyApplyLabel, "label2"},
		},
		Entry{
			Property{PropertyFrom, "foobar@mail.com"},
			Property{PropertyHas, `{pippo "pluto paperino"}`},
			Property{PropertyApplyLabel, "label3"},
		},
	}
	assert.Equal(t, expected, entries)
}

func TestGenerateConsts(t *testing.T) {
	// Test constants replacement
	mf := MatchFilters{
		From: []string{"friends"},
	}
	actions := Actions{
		MarkImportant: true,
	}
	config := Config{
		Consts: Consts{
			"friends": ConstValue{Values: []string{"a@b.com", "b@c.it"}},
			"spam":    ConstValue{Values: []string{"a@spam.com"}},
			"foo":     ConstValue{Values: []string{"useless"}},
		},
		Rules: []Rule{{ /* single empty rule */ }},
	}
	config.Rules[0].Filters.Consts.MatchFilters = mf
	config.Rules[0].Actions = actions

	entries, err := GenerateRules(config)
	assert.Nil(t, err)
	expected := []Entry{
		Entry{
			Property{PropertyFrom, "{a@b.com b@c.it}"},
			Property{PropertyMarkImportant, "true"},
		},
	}
	assert.Equal(t, expected, entries)

	// Test multiple constants in the same clause
	mf = MatchFilters{
		From: []string{"friends", "spam"},
	}
	config.Rules[0].Filters.Consts.MatchFilters = mf
	entries, err = GenerateRules(config)
	assert.Nil(t, err)
	expected = []Entry{
		Entry{
			Property{PropertyFrom, "{a@b.com b@c.it a@spam.com}"},
			Property{PropertyMarkImportant, "true"},
		},
	}
	assert.Equal(t, expected, entries)

	// Test constants in multiple clauses
	mf = MatchFilters{
		From: []string{"friends"},
		To:   []string{"spam"},
	}
	config.Rules[0].Filters.Consts.MatchFilters = mf
	entries, err = GenerateRules(config)
	assert.Nil(t, err)
	expected = []Entry{
		Entry{
			Property{PropertyFrom, "{a@b.com b@c.it}"},
			Property{PropertyTo, "a@spam.com"},
			Property{PropertyMarkImportant, "true"},
		},
	}
	assert.Equal(t, expected, entries)

	// Test unknown constant
	mf = MatchFilters{
		From: []string{"wtf"},
	}
	config.Rules[0].Filters.Consts.MatchFilters = mf
	_, err = GenerateRules(config)
	assert.NotNil(t, err)
}

func TestGenerateNot(t *testing.T) {
	// Test constants replacement
	mf := MatchFilters{
		To:  []string{"my@self.com"},
		Has: []string{"foo", "bar baz"},
	}
	actions := Actions{
		MarkImportant: true,
	}
	config := Config{
		Rules: []Rule{{ /* single empty rule */ }},
	}
	config.Rules[0].Filters.Not = mf
	config.Rules[0].Actions = actions

	entries, err := GenerateRules(config)
	assert.Nil(t, err)
	expected := []Entry{
		Entry{
			Property{PropertyHas, `-{to:my@self.com} -{foo "bar baz"}`},
			Property{PropertyMarkImportant, "true"},
		},
	}
	assert.Equal(t, expected, entries)
}

func TestGenerateNotConsts(t *testing.T) {
	// Test constants replacement
	mf := MatchFilters{
		From: []string{"friends"},
		Has:  []string{"foo"},
	}
	actions := Actions{
		MarkImportant: true,
	}
	config := Config{
		Consts: Consts{
			"friends": ConstValue{Values: []string{"a@b.com", "b@c.it"}},
			"foo":     ConstValue{Values: []string{"useless stuff"}},
		},
		Rules: []Rule{{ /* single empty rule */ }},
	}
	config.Rules[0].Filters.Consts.Not = mf
	config.Rules[0].Actions = actions

	entries, err := GenerateRules(config)
	assert.Nil(t, err)
	expected := []Entry{
		Entry{
			Property{PropertyHas, `-{from:{a@b.com b@c.it}} -"useless stuff"`},
			Property{PropertyMarkImportant, "true"},
		},
	}
	assert.Equal(t, expected, entries)
}
