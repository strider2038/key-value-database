package wal

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/spf13/afero"
	"github.com/strider2038/key-value-database/internal/database/querylang"
)

// LSN - Log Sequence Number, уникальный идентификатор записи в WAL журнале.
// Состоит из двух частей. SessionID - идентификатор сеанса работы сервера, который
// генерируется как метка времени запуска. SeqID - идентификатор выполняемой операции,
// генерируется как последовательное число.
type LSN struct {
	SessionID uint64
	SeqID     uint64
}

func (lsn LSN) Compare(compared LSN) int {
	if lsn.SessionID < compared.SessionID {
		return -1
	}
	if lsn.SessionID > compared.SessionID {
		return 1
	}
	if lsn.SeqID < compared.SeqID {
		return -1
	}
	if lsn.SeqID > compared.SeqID {
		return 1
	}

	return 0
}

type LogRecord struct {
	LSN       LSN
	CommandID querylang.CommandID
	Arguments []string
}

type LogTask struct {
	Record *LogRecord
	Err    chan error
}

// Log - сервис для работы WAL журналом. Обеспечивает операции добавления команд в журнал,
// их извлечение и процедуру обслуживания.
type Log struct {
	reader *Reader
	writer *Writer
	logger *slog.Logger

	flushingBatchSize    int
	flushingBatchTimeout time.Duration

	sessionID uint64

	mu     sync.Mutex
	buffer []*LogTask
	queue  chan []*LogTask
}

func NewLog(
	fs afero.Fs,
	logger *slog.Logger,
	flushingBatchSize int,
	flushingBatchTimeout time.Duration,
	maxSegmentSize int,
	dataDirectory string,
) (*Log, error) {
	if flushingBatchSize <= 0 {
		return nil, fmt.Errorf("flushing batch size must be > 0")
	}
	if flushingBatchTimeout <= 0 {
		return nil, fmt.Errorf("flushing batch timeout must be > 0")
	}

	sessionID := uint64(time.Now().UTC().UnixMilli())
	writer, err := NewWriter(fs, logger, maxSegmentSize, dataDirectory, sessionID)
	if err != nil {
		return nil, err
	}

	return &Log{
		reader:               NewReader(fs, dataDirectory),
		writer:               writer,
		logger:               logger,
		flushingBatchSize:    flushingBatchSize,
		flushingBatchTimeout: flushingBatchTimeout,
		sessionID:            sessionID,
		queue:                make(chan []*LogTask),
	}, nil
}

// Add - добавляет команду в журнал WAL. Команда отправляется сначала в буфер команд.
// Сброс команд из буфера в журнал записи осуществляется по достижении лимита
// flushingBatchSize или по срабатыванию таймера flushingBatchTimeout.
// Операция возвращает управление только после записи всех данных на жесткий диск.
func (l *Log) Add(command *querylang.Command) error {
	start := time.Now()

	task := &LogTask{
		Record: &LogRecord{
			LSN: LSN{
				SessionID: l.sessionID,
				SeqID:     command.SeqID(),
			},
			CommandID: command.ID(),
			Arguments: command.Arguments(),
		},
		Err: make(chan error),
	}

	l.withLock(func() {
		l.buffer = append(l.buffer, task)
		if len(l.buffer) >= l.flushingBatchSize {
			// если буфер заполнился, то сразу сбрасываем его
			l.flush()
		} else if len(l.buffer) == 1 {
			// по добавлению первого элемента в буфер (если максимальный размер != 1),
			// запускаем таймер на сброс буфера
			l.flushByTimeout()
		}
	})

	err := <-task.Err
	if err == nil {
		l.logger.Debug(
			"command added to WAL",
			slog.Uint64("sessionID", l.sessionID),
			slog.Uint64("seqID", command.SeqID()),
			slog.Duration("duration", time.Since(start)),
		)
	}

	return err
}

// Serve - сервисная функция для обслуживания WAL журнала. Ее необходимо запускать
// в фоне работы приложения для корректной работы журнала.
// Функция обеспечивает периодический сброс накопленных команд на жесткий диск.
// Завершается по получению сигнала отмены контекста.
func (l *Log) Serve(ctx context.Context) {
	waiter := sync.WaitGroup{}
	waiter.Add(2)
	go func() {
		defer waiter.Done()
		<-ctx.Done()
		close(l.queue)
	}()
	go func() {
		defer waiter.Done()
		l.serveQueue()
	}()
	waiter.Wait()
}

// Restore - восстанавливает команды из WAL журнала.
func (l *Log) Restore() ([]*querylang.Command, error) {
	records, err := l.reader.ReadRecords()
	if err != nil {
		return nil, err
	}

	commands := make([]*querylang.Command, 0, len(records))
	for _, record := range records {
		commands = append(commands, querylang.NewCommand(record.LSN.SeqID, record.CommandID, record.Arguments...))
	}

	return commands, nil
}

func (l *Log) flushByTimeout() {
	timer := time.NewTimer(l.flushingBatchTimeout)
	defer timer.Stop()

	<-timer.C
	l.flush()
}

// flush запускает процедуру записи накопленного буфер команд на жесткий диск.
// Накопленный буфер buffer отправляется в отдельный канал queue и очищается.
// Отправка через канал делается для минимизации блокировки мьютекса для буфера.
func (l *Log) flush() {
	if len(l.buffer) > 0 {
		l.queue <- l.buffer
		l.buffer = nil
	}
}

// serveQueue обслуживает канал очереди записи команд в WAL журнал.
func (l *Log) serveQueue() {
	for tasks := range l.queue {
		records := make([]*LogRecord, 0, len(tasks))
		for _, task := range tasks {
			records = append(records, task.Record)
		}

		err := l.writer.WriteRecords(records)
		if err != nil {
			for _, task := range tasks {
				task.Err <- err
			}
		} else {
			for _, task := range tasks {
				close(task.Err)
			}
		}
	}
}

func (l *Log) withLock(f func()) {
	l.mu.Lock()
	defer l.mu.Unlock()
	f()
}
