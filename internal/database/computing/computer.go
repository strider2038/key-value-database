package computing

import (
	"context"
	"fmt"

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
}

func NewComputer(parser Parser, analyzer Analyzer) *Computer {
	return &Computer{parser: parser, analyzer: analyzer}
}

func (c *Computer) ParseRequest(ctx context.Context, request string) (*querylang.Command, error) {
	return nil, fmt.Errorf("not implemented")
}
