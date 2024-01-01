package di

import (
	"io"
	"log/slog"
	"os"

	"github.com/strider2038/key-value-database/internal/config"
	"github.com/strider2038/key-value-database/internal/database"
	"github.com/strider2038/key-value-database/internal/database/computation/basic"
	"github.com/strider2038/key-value-database/internal/database/computation/basic/analyzing"
	"github.com/strider2038/key-value-database/internal/database/computation/basic/parsing"
	"github.com/strider2038/key-value-database/internal/database/engine"
	"github.com/strider2038/key-value-database/internal/database/network"
	"github.com/strider2038/key-value-database/internal/database/storage/inmemory"
)

func NewServer(options config.ServerOptions) (*database.Server, error) {
	logger := newLogger(options.Logging)

	tcpServer := network.NewTCPServer(
		options.Network.Address,
		options.Network.MaxConnections,
		options.Network.MaxMessageSize,
		options.Network.IdleTimeout,
		options.Network.OnServerStart,
		logger,
	)

	controller := engine.NewController(
		basic.NewComputer(parsing.NewParser(), analyzing.NewAnalyzer(), logger),
		inmemory.NewMapStorage(),
		logger,
	)
	server := database.NewServer(
		controller,
		tcpServer,
	)

	return server, nil
}

func newLogger(logging config.Logging) *slog.Logger {
	var handler slog.Handler

	options := &slog.HandlerOptions{Level: parseLogLevel(logging.Level)}

	switch logging.Output {
	case "", "discard":
		handler = slog.NewTextHandler(io.Discard, options)
	case "stdout":
		handler = slog.NewTextHandler(os.Stdout, options)
	case "stderr":
		handler = slog.NewTextHandler(os.Stderr, options)
	default:
		// todo: file handler
	}

	return slog.New(handler)
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
