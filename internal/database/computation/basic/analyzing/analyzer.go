package analyzing

import (
	"fmt"

	"github.com/strider2038/key-value-database/internal/database/querylang"
)

type Analyzer struct{}

func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

func (a *Analyzer) AnalyzeCommand(tokens []string) (*querylang.Command, error) {
	if len(tokens) == 0 {
		return nil, ErrEmptyTokens
	}

	commandID := tokens[0]
	arguments := tokens[1:]

	switch commandID {
	case "GET":
		return newCommand(querylang.CommandGet, 1, arguments)
	case "SET":
		return newCommand(querylang.CommandSet, 2, arguments)
	case "DEL":
		return newCommand(querylang.CommandDel, 1, arguments)
	}

	return nil, ErrUnknownCommand
}

func newCommand(id querylang.CommandID, argumentsCount int, arguments []string) (*querylang.Command, error) {
	if len(arguments) < argumentsCount {
		return nil, fmt.Errorf("invalid %q command: %w", id, ErrNotEnoughArguments)
	}
	if len(arguments) > argumentsCount {
		return nil, fmt.Errorf("invalid %q command: %w", id, ErrTooMuchArguments)
	}

	return querylang.NewCommand(id, arguments...), nil
}
