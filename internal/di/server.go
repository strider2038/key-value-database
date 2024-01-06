package di

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

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
	logger, err := newLogger(options.Logging)
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}

	tcpServer, err := network.NewTCPServer(
		options.Network.Address,
		options.Network.MaxConnections,
		options.Network.MaxMessageSize,
		options.Network.IdleTimeout,
		options.Network.OnServerStart,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("create TCP server: %w", err)
	}

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

func newLogger(logging config.Logging) (*slog.Logger, error) {
	var output io.Writer

	switch logging.Output {
	case "", "discard":
		output = io.Discard
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		file, err := createLogFile(logging.Output)
		if err != nil {
			return nil, fmt.Errorf("create log file: %w", err)
		}
		output = file
	}

	options := &slog.HandlerOptions{Level: parseLogLevel(logging.Level)}

	return slog.New(slog.NewTextHandler(output, options)), nil
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

func createLogFile(filename string) (*os.File, error) {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("create log dir %q: %w", dir, err)
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return file, nil
}
