package basic

import (
	"fmt"
	"log/slog"

	"github.com/strider2038/key-value-database/internal/database/querylang"
)

type Parser interface {
	ParseCommand(command string) ([]string, error)
}

type Analyzer interface {
	AnalyzeCommand(tokens []string) (*querylang.Command, error)
}

type Computer struct {
	parser   Parser
	analyzer Analyzer
	logger   *slog.Logger
}

func NewComputer(parser Parser, analyzer Analyzer, logger *slog.Logger) *Computer {
	return &Computer{parser: parser, analyzer: analyzer, logger: logger}
}

func (c *Computer) ParseRequest(request string) (*querylang.Command, error) {
	tokens, err := c.parser.ParseCommand(request)
	if err != nil {
		return nil, fmt.Errorf("parse command: %w", err)
	}

	c.logger.With("tokens", tokens).Debug("parsing tokens completed")

	command, err := c.analyzer.AnalyzeCommand(tokens)
	if err != nil {
		return nil, fmt.Errorf("analyze command: %w", err)
	}

	return command, nil
}
