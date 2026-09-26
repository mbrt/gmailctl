package cmd

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAskYN(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    bool
		wantOut string
	}{
		{"yes", "y\n", true, "Continue? [y/N]: "},
		{"yes uppercase with CRLF", "YES\r\n", true, "Continue? [y/N]: "},
		{"no", "n\n", false, "Continue? [y/N]: "},
		{"empty line defaults to no", "\n", false, "Continue? [y/N]: "},
		{"retry after invalid choice", "maybe\ny\n", true,
			"Continue? [y/N]: invalid choice\nContinue? [y/N]: "},
		// https://github.com/mbrt/gmailctl/issues/464
		{"EOF defaults to no", "", false, "Continue? [y/N]: \n"},
		{"EOF after invalid choice defaults to no", "maybe\n", false,
			"Continue? [y/N]: invalid choice\nContinue? [y/N]: \n"},
		{"unterminated answer before EOF", "y", true, "Continue? [y/N]: \n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out strings.Builder
			got := askYN(strings.NewReader(tt.input), &out, "Continue?")
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantOut, out.String())
		})
	}
}

func TestAskOptions(t *testing.T) {
	choices := []string{"yes", "no", "abort"}
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"full choice", "abort\n", 2},
		{"prefix uppercase", "Y\n", 0},
		{"retry after empty line", "\nn\n", 1},
		{"retry after invalid choice", "x\na\n", 2},
		{"unterminated answer before EOF", "a", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := askOptions(strings.NewReader(tt.input), io.Discard, "Apply?", choices)
			require.Nil(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// https://github.com/mbrt/gmailctl/issues/464
func TestAskOptionsEOF(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"no input", ""},
		{"EOF after invalid choice", "x\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := askOptions(strings.NewReader(tt.input), io.Discard, "Apply?", []string{"yes", "no"})
			assert.ErrorIs(t, err, io.EOF)
		})
	}
}
