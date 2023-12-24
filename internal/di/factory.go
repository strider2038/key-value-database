package di

import (
	"io"
	"log/slog"

	"github.com/strider2038/key-value-database/internal/database"
	"github.com/strider2038/key-value-database/internal/database/computation/basic"
	"github.com/strider2038/key-value-database/internal/database/computation/basic/analyzing"
	"github.com/strider2038/key-value-database/internal/database/computation/basic/parsing"
	"github.com/strider2038/key-value-database/internal/database/storage/inmemory"
)

type Options struct {
	Logger *slog.Logger
}

func NewDatabase(options Options) *database.Database {
	if options.Logger == nil {
		options.Logger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	}

	return database.NewDatabase(
		basic.NewComputer(parsing.NewParser(), analyzing.NewAnalyzer(), options.Logger),
		inmemory.NewMapStorage(),
		options.Logger,
	)
}
