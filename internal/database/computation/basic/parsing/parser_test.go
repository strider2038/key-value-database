package parsing_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/strider2038/key-value-database/internal/database/computation/basic/parsing"
)

func TestParser_ParseCommand(t *testing.T) {
	tests := []struct {
		input      string
		wantTokens []string
		wantError  error
	}{
		{
			input:      "",
			wantTokens: []string{},
		},
		{
			input:      "GET token",
			wantTokens: []string{"GET", "token"},
		},
		{
			input:      "foo bar baz",
			wantTokens: []string{"foo", "bar", "baz"},
		},
		{
			input:      " foo\tbar\r\nbaz ",
			wantTokens: []string{"foo", "bar", "baz"},
		},
		{
			input:      "digits123 letters punctuation*/_",
			wantTokens: []string{"digits123", "letters", "punctuation*/_"},
		},
		{
			input:     "unexpected =",
			wantError: parsing.ErrUnexpectedSymbol,
		},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			parser := parsing.NewParser()

			gotTokens, gotErr := parser.ParseCommand(test.input)

			if test.wantError == nil {
				require.NoError(t, gotErr)
				assert.Equal(t, test.wantTokens, gotTokens)
			} else {
				assert.Nil(t, gotTokens)
				assert.ErrorIs(t, gotErr, test.wantError)
			}
		})
	}
}
