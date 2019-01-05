package v2alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "github.com/mbrt/gmailctl/pkg/config/v1alpha1"
)

func v1Config(c v1.Config) v1.Config {
	c.Version = v1.Version
	return c
}

func v2Config(c Config) Config {
	c.Version = Version
	return c
}

func TestConvert(t *testing.T) {
	tests := []struct {
		name string
		cfg  v1.Config
		exp  Config
	}{
		{
			"empty",
			v1Config(v1.Config{}),
			v2Config(Config{}),
		},
		{
			"preserved author",
			v1Config(v1.Config{
				Author: v1.Author{Name: "foo", Email: "bar"},
			}),
			v2Config(Config{
				Author: Author{Name: "foo", Email: "bar"},
			}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Import(tc.cfg)
			assert.Nil(t, err)
			assert.Equal(t, tc.exp, got)
		})
	}
}
