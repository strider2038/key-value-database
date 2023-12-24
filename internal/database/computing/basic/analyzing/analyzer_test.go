package analyzing_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/strider2038/key-value-database/internal/database/computing/basic/analyzing"
	"github.com/strider2038/key-value-database/internal/database/querylang"
)

func TestAnalyzer_AnalyzeCommand(t *testing.T) {
	tests := []struct {
		name          string
		tokens        []string
		wantCommand   querylang.CommandID
		wantArguments []string
		wantError     error
	}{
		{
			name:      "empty input",
			tokens:    []string{},
			wantError: analyzing.ErrEmptyTokens,
		},
		{
			name:      "unknown command",
			tokens:    strings.Fields("unknown"),
			wantError: analyzing.ErrUnknownCommand,
		},
		{
			name:          "get command: valid",
			tokens:        strings.Fields("GET key1"),
			wantCommand:   querylang.CommandGet,
			wantArguments: []string{"key1"},
		},
		{
			name:      "get command: not enough arguments",
			tokens:    strings.Fields("GET"),
			wantError: analyzing.ErrNotEnoughArguments,
		},
		{
			name:      "get command: too much arguments",
			tokens:    strings.Fields("GET key1 key2"),
			wantError: analyzing.ErrTooMuchArguments,
		},
		{
			name:          "set command: valid",
			tokens:        strings.Fields("SET key1 value1"),
			wantCommand:   querylang.CommandSet,
			wantArguments: []string{"key1", "value1"},
		},
		{
			name:      "set command: not enough arguments",
			tokens:    strings.Fields("SET key1"),
			wantError: analyzing.ErrNotEnoughArguments,
		},
		{
			name:      "set command: too much arguments",
			tokens:    strings.Fields("SET key1 key2 key3"),
			wantError: analyzing.ErrTooMuchArguments,
		},
		{
			name:          "del command: valid",
			tokens:        strings.Fields("DEL key1"),
			wantCommand:   querylang.CommandDel,
			wantArguments: []string{"key1"},
		},
		{
			name:      "del command: not enough arguments",
			tokens:    strings.Fields("DEL"),
			wantError: analyzing.ErrNotEnoughArguments,
		},
		{
			name:      "del command: too much arguments",
			tokens:    strings.Fields("DEL key1 key2"),
			wantError: analyzing.ErrTooMuchArguments,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			analyzer := analyzing.NewAnalyzer()

			command, err := analyzer.AnalyzeCommand(test.tokens)

			if test.wantError == nil {
				require.NoError(t, err)
				assert.Equal(t, test.wantCommand.String(), command.ID().String())
				assert.Equal(t, test.wantArguments, command.Arguments())
			} else {
				assert.Nil(t, command)
				assert.ErrorIs(t, err, test.wantError)
			}
		})
	}
}
