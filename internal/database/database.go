package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/strider2038/key-value-database/internal/database/querylang"
	"github.com/strider2038/key-value-database/internal/database/storage"
)

const Nil = "$_"

type Computer interface {
	ParseRequest(request string) (*querylang.Command, error)
}

type Storage interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	Del(ctx context.Context, key string) error
}

type Database struct {
	computer Computer
	storage  Storage
	logger   *slog.Logger
}

func NewDatabase(
	computer Computer,
	storage Storage,
	logger *slog.Logger,
) *Database {
	return &Database{
		computer: computer,
		storage:  storage,
		logger:   logger,
	}
}

func (db *Database) Execute(ctx context.Context, request string) (string, error) {
	command, err := db.parseCommand(request)
	if err != nil {
		return "", err
	}

	result, err := db.handleCommand(ctx, command)
	if err != nil {
		return "", fmt.Errorf("handle %s command: %w", command.ID(), err)
	}

	return result, nil
}

func (db *Database) parseCommand(request string) (*querylang.Command, error) {
	start := time.Now()

	command, err := db.computer.ParseRequest(request)
	if err != nil {
		return nil, fmt.Errorf("parse request: %w", err)
	}

	db.logger.
		With(
			slog.String("request", request),
			slog.Duration("duration", time.Since(start)),
			slog.String("commandID", command.ID().String()),
			slog.Any("commandArgs", command.Arguments()),
		).
		Debug("command parsing completed")

	return command, nil
}

func (db *Database) handleCommand(ctx context.Context, command *querylang.Command) (string, error) {
	start := time.Now()
	defer func() {
		db.logger.
			With(
				slog.Duration("duration", time.Since(start)),
				slog.String("commandID", command.ID().String()),
				slog.Any("commandArgs", command.Arguments()),
			).
			Debug("command execution completed")
	}()

	switch command.ID() {
	case querylang.CommandGet:
		return db.handleGet(ctx, command.Arguments())
	case querylang.CommandSet:
		return db.handleSet(ctx, command.Arguments())
	case querylang.CommandDel:
		return db.handleDel(ctx, command.Arguments())
	default:
		db.logger.Error("unsupported command", "commandID", command.ID().String())

		return "", nil
	}
}

func (db *Database) handleGet(ctx context.Context, arguments []string) (string, error) {
	value, err := db.storage.Get(ctx, arguments[0])
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return Nil, nil
		}

		return "", err
	}

	return value, nil
}

func (db *Database) handleSet(ctx context.Context, arguments []string) (string, error) {
	if err := db.storage.Set(ctx, arguments[0], arguments[1]); err != nil {
		return "", err
	}

	return "OK", nil
}

func (db *Database) handleDel(ctx context.Context, arguments []string) (string, error) {
	if err := db.storage.Del(ctx, arguments[0]); err != nil {
		return "", err
	}

	return "OK", nil
}
