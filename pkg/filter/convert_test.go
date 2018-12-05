package filter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mbrt/gmailctl/pkg/config"
)

func TestMatchFilter(t *testing.T) {
	// Test a single filter
	filt := config.MatchFilters{
		Subject: []string{"important", "not important"},
	}
	crit := generateMatchFilters(filt)
	expected := Criteria{
		Subject: `{important "not important"}`,
	}
	assert.Equal(t, expected, crit)

	// Test all the filters together
	filt = config.MatchFilters{
		From:    []string{"foobar@mail.com", "baz@g.com"},
		To:      []string{"my@self.com"},
		Cc:      []string{"other@self.com"},
		Subject: []string{"important", "not important"},
		Has:     []string{"what's wrong", "alert"},
		List:    []string{"wow-list@l.com"},
	}
	crit = generateMatchFilters(filt)
	expected = Criteria{
		From:    "{foobar@mail.com baz@g.com}",
		To:      "my@self.com",
		Subject: `{important "not important"}`,
		Query:   `{"what's wrong" alert} list:wow-list@l.com cc:other@self.com`,
	}
	assert.Equal(t, expected, crit)
}

func TestNotFilter(t *testing.T) {
	// Test a single filter
	filt := config.MatchFilters{
		From:    []string{"foobar@mail.com", "baz@g.com"},
		To:      []string{"my@self.com"},
		Cc:      []string{"other@self.com"},
		Subject: []string{"important", "not important"},
		Has:     []string{"what's wrong", "alert"},
		List:    []string{"wow-list@l.com"},
	}
	crit := generateNegatedFilters(filt)
	expected := strings.Join([]string{ // for readability
		`-{from:{foobar@mail.com baz@g.com}}`,
		`-{to:my@self.com}`,
		`-{cc:other@self.com}`,
		`-{subject:{important "not important"}}`,
		`-{"what's wrong" alert}`,
		`-{list:wow-list@l.com}`,
	}, " ")
	assert.Equal(t, expected, crit)
}

func TestCombineMatchAndNegated(t *testing.T) {
	// Test combining a positive with a negative filter
	filt := config.Filters{
		CompositeFilters: config.CompositeFilters{
			MatchFilters: config.MatchFilters{
				From: []string{"*@mail.com"},
				Has:  []string{"zumba"},
			},
			Not: config.MatchFilters{
				From: []string{"baz@mail.com"},
			},
		},
	}
	crit := generateCriteria(filt)
	expected := Criteria{
		From:  "*@mail.com",
		Query: "zumba -{from:baz@mail.com}",
	}
	assert.Equal(t, expected, crit)
}

func TestList(t *testing.T) {
	filt := config.Filters{
		CompositeFilters: config.CompositeFilters{
			MatchFilters: config.MatchFilters{
				List: []string{"list@mail.com"},
			},
		},
	}
	crit := generateCriteria(filt)
	expected := Criteria{
		Query: "list:list@mail.com",
	}
	assert.Equal(t, expected, crit)
}

func TestCombineWithQuery(t *testing.T) {
	// Test combining custom query with other filters
	filt := config.Filters{
		CompositeFilters: config.CompositeFilters{
			MatchFilters: config.MatchFilters{
				From: []string{"*@mail.com"},
				List: []string{"list@mail.com"},
			},
			Not: config.MatchFilters{
				From: []string{"baz@mail.com"},
			},
		},
		Query: "foo {bar baz}",
	}
	crit := generateCriteria(filt)
	expected := Criteria{
		From:  "*@mail.com",
		Query: "list:list@mail.com -{from:baz@mail.com} foo {bar baz}",
	}
	assert.Equal(t, expected, crit)
}

func TestActions(t *testing.T) {
	// Test all the actions together
	act := config.Actions{
		Archive:       true,
		Delete:        true,
		MarkImportant: true,
		MarkRead:      true,
		Category:      config.CategoryPersonal,
		Labels:        []string{"label1", "label2"},
	}
	props := generateActions(act)
	expected := []Action{
		{
			Archive:       true,
			Delete:        true,
			MarkImportant: true,
			MarkRead:      true,
			Category:      config.CategoryPersonal,
			AddLabel:      "label1",
		},
		{
			AddLabel: "label2",
		},
	}
	assert.Equal(t, expected, props)
}

func TestGenerateSingleEntry(t *testing.T) {
	// Smoke test with a single entry as result
	mf := config.MatchFilters{
		From: []string{"foobar@mail.com"},
	}
	actions := config.Actions{
		Archive:  true,
		MarkRead: true,
	}
	cfg := config.Config{
		Rules: []config.Rule{{ /* single empty rule */ }},
	}
	cfg.Rules[0].Filters.MatchFilters = mf
	cfg.Rules[0].Actions = actions

	entries, err := FromConfig(cfg)
	assert.Nil(t, err)
	expected := Filters{
		{
			Criteria: Criteria{
				From: "foobar@mail.com",
			},
			Action: Action{
				Archive:  true,
				MarkRead: true,
			},
		},
	}
	assert.Equal(t, expected, entries)
}

func TestGenerateMultipleEntities(t *testing.T) {
	// Smoke test with a single entry as result
	mf := config.MatchFilters{
		From: []string{"foobar@mail.com"},
		Has:  []string{"pippo", "pluto paperino"},
	}
	actions := config.Actions{
		MarkRead: true,
		Labels:   []string{"label1", "label2", "label3"},
	}
	config := config.Config{
		Rules: []config.Rule{{ /* single empty rule */ }},
	}
	config.Rules[0].Filters.MatchFilters = mf
	config.Rules[0].Actions = actions

	entries, err := FromConfig(config)
	assert.Nil(t, err)
	expected := Filters{
		{
			Criteria: Criteria{
				From:  "foobar@mail.com",
				Query: `{pippo "pluto paperino"}`,
			},
			Action: Action{
				MarkRead: true,
				AddLabel: "label1",
			},
		},
		{
			Criteria: Criteria{
				From:  "foobar@mail.com",
				Query: `{pippo "pluto paperino"}`,
			},
			Action: Action{
				AddLabel: "label2",
			},
		},
		{
			Criteria: Criteria{
				From:  "foobar@mail.com",
				Query: `{pippo "pluto paperino"}`,
			},
			Action: Action{
				AddLabel: "label3",
			},
		},
	}
	assert.Equal(t, expected, entries)
}

func TestGenerateConsts(t *testing.T) {
	// Test constants replacement
	mf := config.MatchFilters{
		From: []string{"friends"},
	}
	actions := config.Actions{
		MarkImportant: true,
	}
	cfg := config.Config{
		Consts: config.Consts{
			"friends": config.ConstValue{Values: []string{"a@b.com", "b@c.it"}},
			"spam":    config.ConstValue{Values: []string{"a@spam.com"}},
			"foo":     config.ConstValue{Values: []string{"useless"}},
		},
		Rules: []config.Rule{{ /* single empty rule */ }},
	}
	cfg.Rules[0].Filters.Consts.MatchFilters = mf
	cfg.Rules[0].Actions = actions

	entries, err := FromConfig(cfg)
	assert.Nil(t, err)
	expected := Filters{
		{
			Criteria: Criteria{
				From: "{a@b.com b@c.it}",
			},
			Action: Action{
				MarkImportant: true,
			},
		},
	}
	assert.Equal(t, expected, entries)

	// Test multiple constants in the same clause
	mf = config.MatchFilters{
		From: []string{"friends", "spam"},
	}
	cfg.Rules[0].Filters.Consts.MatchFilters = mf
	entries, err = FromConfig(cfg)
	assert.Nil(t, err)
	expected = Filters{
		{
			Criteria: Criteria{
				From: "{a@b.com b@c.it a@spam.com}",
			},
			Action: Action{
				MarkImportant: true,
			},
		},
	}
	assert.Equal(t, expected, entries)

	// Test constants in multiple clauses
	mf = config.MatchFilters{
		From: []string{"friends"},
		To:   []string{"spam"},
	}
	cfg.Rules[0].Filters.Consts.MatchFilters = mf
	entries, err = FromConfig(cfg)
	assert.Nil(t, err)
	expected = Filters{
		{
			Criteria: Criteria{
				From: "{a@b.com b@c.it}",
				To:   "a@spam.com",
			},
			Action: Action{
				MarkImportant: true,
			},
		},
	}
	assert.Equal(t, expected, entries)

	// Test unknown constant
	mf = config.MatchFilters{
		From: []string{"wtf"},
	}
	cfg.Rules[0].Filters.Consts.MatchFilters = mf
	_, err = FromConfig(cfg)
	assert.NotNil(t, err)
}
