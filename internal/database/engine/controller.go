package engine

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/strider2038/key-value-database/internal/database/computation"
	"github.com/strider2038/key-value-database/internal/database/querylang"
)

type Computer interface {
	ParseRequest(request string) (*computation.Command, error)
}

type StorageController interface {
	Execute(command *querylang.Command) (string, error)
}

type Controller struct {
	computer          Computer
	storageController StorageController
	idGenerator       IDGenerator
	logger            *slog.Logger
}

func NewController(
	computer Computer,
	storageController StorageController,
	logger *slog.Logger,
) *Controller {
	return &Controller{
		computer:          computer,
		storageController: storageController,
		logger:            logger,
	}
}

func (c *Controller) Execute(ctx context.Context, rawCommand string) (string, error) {
	start := time.Now()

	command, err := c.parseCommand(rawCommand)
	if err != nil {
		return "", &BadRequestError{err: err}
	}

	result, err := c.storageController.Execute(command)
	if err != nil {
		c.logger.Error("command execution failed", "seqID", command.SeqID(), "error", err)

		return "", fmt.Errorf("handle %s command: %w", command.ID(), err)
	}

	c.logger.Info(
		"command execution completed",
		slog.Uint64("seqID", command.SeqID()),
		slog.Duration("duration", time.Since(start)),
		slog.String("commandID", command.ID().String()),
		slog.Any("commandArgs", command.Arguments()),
	)

	return result, nil
}

func (c *Controller) parseCommand(rawCommand string) (*querylang.Command, error) {
	start := time.Now()

	parsedCommand, err := c.computer.ParseRequest(rawCommand)
	if err != nil {
		return nil, fmt.Errorf("parse command: %w", err)
	}

	seqID := c.idGenerator.NextSeqID()
	command := querylang.NewCommand(seqID, parsedCommand.ID, parsedCommand.Arguments...)

	c.logger.
		With(
			slog.Uint64("seqID", command.SeqID()),
			slog.String("rawCommand", rawCommand),
			slog.Duration("duration", time.Since(start)),
			slog.String("commandID", command.ID().String()),
			slog.Any("commandArgs", command.Arguments()),
		).
		Debug("command parsing completed")

	return command, nil
}
