package wal

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/afero"
)

// Writer отвечает за запись элементов WAL журнала в файлы.
type Writer struct {
	fs     afero.Fs
	logger *slog.Logger

	file afero.File

	sessionID      uint64
	segmentNo      int
	segmentSize    int
	maxSegmentSize int
	directory      string
}

func NewWriter(
	fs afero.Fs,
	logger *slog.Logger,
	maxSegmentSize int,
	directory string,
	sessionID uint64,
) (*Writer, error) {
	if maxSegmentSize <= 0 {
		return nil, fmt.Errorf("segment size must be > 0")
	}

	return &Writer{
		fs:             fs,
		logger:         logger,
		maxSegmentSize: maxSegmentSize,
		directory:      strings.TrimSuffix(directory, "/"),
		sessionID:      sessionID,
	}, nil
}

// WriteRecords записывает пачку элементов журнала в текущий файл сегмента.
// Если сегмента еще не существует или достигнут лимит размера maxSegmentSize,
// то осуществляет ротацию сегмента на новый файл.
func (w *Writer) WriteRecords(records []*LogRecord) error {
	start := time.Now()

	if w.file == nil || w.segmentSize > w.maxSegmentSize {
		if err := w.rotate(); err != nil {
			return fmt.Errorf("rotate WAL file: %w", err)
		}
	}

	if err := w.write(records); err != nil {
		return fmt.Errorf("write to WAL file: %w", err)
	}

	// Sync сбрасывает буферы i/o на жесткий диск. Т.о. гарантируется, что
	// в файлы были записаны данные.
	if err := w.file.Sync(); err != nil {
		return fmt.Errorf("sync WAL file: %w", err)
	}

	w.logger.Info(
		"log records written to WAL",
		slog.Int("recordsCount", len(records)),
		slog.Uint64("sessionID", w.sessionID),
		slog.Duration("duration", time.Since(start)),
	)

	return nil
}

// rotate создает новый файл сегмента WAL журнала. Файлы создаются в директории
// directory, имя формируется как wal_<session_id>_<segment_no>.log,
// где session_id - идентификатор сессии WAL журнала, segment_no - последовательный номер
// сегмента из текущей сессии.
func (w *Writer) rotate() error {
	if w.file != nil {
		w.file.Close()
	}

	if err := w.fs.MkdirAll(w.directory, os.ModePerm); err != nil {
		return fmt.Errorf("create WAL directory: %w", err)
	}

	filename := fmt.Sprintf("%s/wal_%d_%08d.log", w.directory, w.sessionID, w.segmentNo)
	file, err := w.fs.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open WAL: %w", err)
	}

	w.file = file
	w.segmentSize = 0
	w.segmentNo++

	w.logger.Info(
		"WAL segment created",
		slog.String("walSegment", filename),
		slog.Uint64("sessionID", w.sessionID),
	)

	return nil
}

func (w *Writer) write(records []*LogRecord) error {
	buffer := bytes.Buffer{}
	if err := gob.NewEncoder(&buffer).Encode(records); err != nil {
		return fmt.Errorf("encode records: %w", err)
	}

	size, err := w.file.Write(buffer.Bytes())
	if err != nil {
		return err
	}

	w.segmentSize += size

	return nil
}
