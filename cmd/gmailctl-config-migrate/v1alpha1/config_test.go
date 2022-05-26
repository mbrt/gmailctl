package v1alpha1

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestParse(t *testing.T) {
	yml := []byte(`
version: v1alpha1
author:
  name: MB
  email: m@gmail.com
rules:
  - filters:
      list:
        - foobar@g.com
    actions:
      labels:
        - MyList
      archive: true
`)
	var res Config
	dec := yaml.NewDecoder(bytes.NewBuffer(yml))
	dec.KnownFields(true)
	assert.NoError(t, dec.Decode(&res))
	filters := Filters{
		CompositeFilters: CompositeFilters{
			MatchFilters: MatchFilters{
				List: []string{"foobar@g.com"},
			},
		},
	}
	expected := Config{
		Version: "v1alpha1",
		Author:  Author{Name: "MB", Email: "m@gmail.com"},
		Rules: []Rule{
			{
				Filters: filters,
				Actions: Actions{
					Labels:  []string{"MyList"},
					Archive: true,
				},
			},
		},
	}
	assert.Equal(t, expected, res)
}

func TestParseUnknownField(t *testing.T) {
	yml := []byte(`
version: v1alpha1
author:
  name: MB
  email: m@gmail.com
rules:
  - filters:
      foobar:
        - foobar@g.com
`)
	var res Config
	dec := yaml.NewDecoder(bytes.NewBuffer(yml))
	dec.KnownFields(true)
	assert.NotNil(t, dec.Decode(&res))
}
