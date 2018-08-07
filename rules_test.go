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
		{PropertyTo, "{my@self.com}"},
		{PropertySubject, `{important "not important"}`},
		{PropertyHas, `{"what's wrong" alert}`},
	}
	assert.Equal(t, expected, props)
}
