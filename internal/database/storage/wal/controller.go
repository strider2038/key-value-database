package wal

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/afero"
	"github.com/strider2038/key-value-database/internal/database/querylang"
)

type StorageController interface {
	Execute(command *querylang.Command) (string, error)
}

// Controller - адаптер контроллера базы данных для работы WAL журнала предзаписи.
type Controller struct {
	storageController StorageController
	log               *Log
	logger            *slog.Logger
}

func NewController(
	storageController StorageController,
	fs afero.Fs,
	logger *slog.Logger,
	flushingBatchSize int,
	flushingBatchTimeout time.Duration,
	maxSegmentSize int,
	dataDirectory string,
) (*Controller, error) {
	log, err := NewLog(fs, logger, flushingBatchSize, flushingBatchTimeout, maxSegmentSize, dataDirectory)
	if err != nil {
		return nil, err
	}
	c := &Controller{
		storageController: storageController,
		log:               log,
		logger:            logger,
	}

	if err := c.restore(); err != nil {
		return nil, fmt.Errorf("restore from WAL: %w", err)
	}

	return c, nil
}

// Execute адаптер для выполнения команд БД. Все операции чтения напрямую делегируются
// нижележащему контроллеру. Операции записи перед выполнением добавляются в WAL журнал.
// Команда записи делегируется нижележащему контроллеру только в случае успешной
// записи в WAL журнал.
func (c *Controller) Execute(command *querylang.Command) (string, error) {
	if command.IsReadOperation() {
		return c.storageController.Execute(command)
	}

	if err := c.log.Add(command); err != nil {
		return "", fmt.Errorf("add to WAL: %w", err)
	}

	return c.storageController.Execute(command)
}

// Serve - сервисная функция для обслуживания WAL журнала. Ее необходимо запускать
// в фоне работы приложения для корректной работы журнала.
// Функция обеспечивает периодический сброс накопленных команд на жесткий диск.
// Завершается по получению сигнала отмены контекста.
func (c *Controller) Serve(ctx context.Context) error {
	c.log.Serve(ctx)

	return nil
}

func (c *Controller) restore() error {
	start := time.Now()

	commands, err := c.log.Restore()
	if err != nil {
		return err
	}

	for _, command := range commands {
		if _, err := c.storageController.Execute(command); err != nil {
			return fmt.Errorf("execute command %s %v", command.ID(), command.Arguments())
		}

		c.logger.Debug(
			"command executed from WAL",
			slog.Uint64("seqID", command.SeqID()),
			slog.String("commandID", command.ID().String()),
			slog.Any("commandArguments", command.Arguments()),
		)
	}

	if len(commands) > 0 {
		c.logger.Info(
			"storage state restored from WAL",
			slog.Int("commandsCount", len(commands)),
			slog.Duration("duration", time.Since(start)),
		)
	} else {
		c.logger.Info("WAL is empty")
	}

	return nil
}
