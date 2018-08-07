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
