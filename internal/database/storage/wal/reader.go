package wal

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/spf13/afero"
)

// Reader - сервис для вычитывания записей из WAL журнала.
type Reader struct {
	fs        afero.Fs
	directory string
}

func NewReader(fs afero.Fs, directory string) *Reader {
	return &Reader{
		fs:        fs,
		directory: strings.TrimSuffix(directory, "/"),
	}
}

// ReadRecords - вычитывает ранее записанные команды из WAL журнала.
// Для этого последовательно читает данные из файлов, а затем сортирует их по меткам LSN.
func (r *Reader) ReadRecords() ([]*LogRecord, error) {
	files, err := afero.ReadDir(r.fs, r.directory)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("read WAL directory: %w", err)
	}

	var records []*LogRecord

	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			continue
		}

		filename := r.directory + "/" + fileInfo.Name()
		segmentRecords, err := r.readSegment(filename)
		if err != nil {
			return nil, fmt.Errorf("read segment: %w", err)
		}
		records = append(records, segmentRecords...)
	}

	slices.SortFunc(records, func(a, b *LogRecord) int {
		return a.LSN.Compare(b.LSN)
	})

	return records, nil
}

func (r *Reader) readSegment(filename string) ([]*LogRecord, error) {
	file, err := r.fs.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("read WAL file %q: %w", filename, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read WAL file contents from %q: %w", filename, err)
	}

	var records []*LogRecord

	buffer := bytes.NewBuffer(data)
	for buffer.Len() > 0 {
		var batch []*LogRecord
		decoder := gob.NewDecoder(buffer)
		if err := decoder.Decode(&batch); err != nil {
			return nil, fmt.Errorf("read WAL records from %q: %w", filename, err)
		}
		records = append(records, batch...)
	}

	return records, nil
}
