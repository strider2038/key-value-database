package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/strider2038/key-value-database/internal/database/querylang"
	"github.com/strider2038/key-value-database/internal/database/storage"
)

type RequestParser interface {
	ParseRequest(request string) (*querylang.Command, error)
}

type Storage interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Del(key string) error
}

type Controller struct {
	requestParser RequestParser
	storage       Storage
	logger        *slog.Logger
}

func NewController(
	requestParser RequestParser,
	storage Storage,
	logger *slog.Logger,
) *Controller {
	return &Controller{
		requestParser: requestParser,
		storage:       storage,
		logger:        logger,
	}
}

func (c *Controller) Execute(ctx context.Context, rawCommand string) (string, error) {
	command, err := c.parseCommand(rawCommand)
	if err != nil {
		return "", &BadRequestError{err: err}
	}

	result, err := c.executeCommand(command)
	if err != nil {
		return "", fmt.Errorf("handle %s command: %w", command.ID(), err)
	}

	return result, nil
}

func (c *Controller) parseCommand(rawCommand string) (*querylang.Command, error) {
	start := time.Now()

	command, err := c.requestParser.ParseRequest(rawCommand)
	if err != nil {
		return nil, fmt.Errorf("parse command: %w", err)
	}

	c.logger.
		With(
			slog.String("rawCommand", rawCommand),
			slog.Duration("duration", time.Since(start)),
			slog.String("commandID", command.ID().String()),
			slog.Any("commandArgs", command.Arguments()),
		).
		Debug("command parsing completed")

	return command, nil
}

func (c *Controller) executeCommand(command *querylang.Command) (string, error) {
	start := time.Now()
	defer func() {
		c.logger.
			With(
				slog.Duration("duration", time.Since(start)),
				slog.String("commandID", command.ID().String()),
				slog.Any("commandArgs", command.Arguments()),
			).
			Debug("command execution completed")
	}()

	switch command.ID() {
	case querylang.CommandGet:
		return c.handleGet(command.Arguments())
	case querylang.CommandSet:
		return c.handleSet(command.Arguments())
	case querylang.CommandDel:
		return c.handleDel(command.Arguments())
	default:
		return "", fmt.Errorf("unsupported command: %s", command.ID().String())
	}
}

func (c *Controller) handleGet(arguments []string) (string, error) {
	value, err := c.storage.Get(arguments[0])
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return querylang.Nil, nil
		}

		return "", err
	}

	return value, nil
}

func (c *Controller) handleSet(arguments []string) (string, error) {
	if err := c.storage.Set(arguments[0], arguments[1]); err != nil {
		return "", err
	}

	return "OK", nil
}

func (c *Controller) handleDel(arguments []string) (string, error) {
	if err := c.storage.Del(arguments[0]); err != nil {
		return "", err
	}

	return "OK", nil
}
