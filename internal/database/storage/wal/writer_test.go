package wal_test

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/strider2038/key-value-database/internal/database/querylang"
	"github.com/strider2038/key-value-database/internal/database/storage/wal"
)

func TestWriter_WriteRecords(t *testing.T) {
	fs := afero.NewMemMapFs()
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	writer, err := wal.NewWriter(fs, logger, 100, walDirectory, uint64(time.Now().UnixMilli()))
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		err = writer.WriteRecords(repeatRecords(
			&wal.LogRecord{
				CommandID: querylang.CommandGet,
				Arguments: []string{"foo_bar_baz"},
			},
			10,
		))
		require.NoError(t, err)
	}

	files, err := afero.ReadDir(fs, walDirectory)
	require.NoError(t, err)
	assert.Greater(t, len(files), 0)
	for _, file := range files {
		assert.Greater(t, int(file.Size()), 0)
	}
	records := readRecords(t, fs)
	for _, record := range records {
		assert.Equal(t, querylang.CommandGet, record.CommandID)
		assert.Equal(t, []string{"foo_bar_baz"}, record.Arguments)
	}
}

func repeatRecords(record *wal.LogRecord, n int) []*wal.LogRecord {
	records := make([]*wal.LogRecord, n)

	for i := 0; i < n; i++ {
		records[i] = &wal.LogRecord{
			CommandID: record.CommandID,
			Arguments: record.Arguments,
		}
	}

	return records
}
