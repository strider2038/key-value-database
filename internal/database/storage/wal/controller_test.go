package wal_test

import (
	"bytes"
	"context"
	"encoding/gob"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/strider2038/key-value-database/internal/database/querylang"
	"github.com/strider2038/key-value-database/internal/database/storage"
	"github.com/strider2038/key-value-database/internal/database/storage/inmemory"
	"github.com/strider2038/key-value-database/internal/database/storage/wal"
)

const walDirectory = "/test/dir"

type Data struct {
	value string
	err   error
}

func TestController_Execute(t *testing.T) {
	tests := []struct {
		name          string
		walRecords    []*wal.LogRecord
		applyCommands []*querylang.Command
		wantData      map[string]Data
		wantRecords   []*wal.LogRecord
	}{
		{
			name: "when empty WAL and apply commands, expect values saved",
			applyCommands: []*querylang.Command{
				querylang.NewCommand(1, querylang.CommandSet, "key1", "foo"),
				querylang.NewCommand(2, querylang.CommandSet, "key2", "bar"),
				querylang.NewCommand(3, querylang.CommandGet, "key1"),
				querylang.NewCommand(4, querylang.CommandDel, "key1"),
			},
			wantData: map[string]Data{
				"key1": {err: storage.ErrNotFound},
				"key2": {value: "bar"},
			},
			wantRecords: []*wal.LogRecord{
				{LSN: wal.LSN{SeqID: 1}, CommandID: querylang.CommandSet, Arguments: []string{"key1", "foo"}},
				{LSN: wal.LSN{SeqID: 2}, CommandID: querylang.CommandSet, Arguments: []string{"key2", "bar"}},
				{LSN: wal.LSN{SeqID: 4}, CommandID: querylang.CommandDel, Arguments: []string{"key1"}},
			},
		},
		{
			name: "when filled WAL and no commands, expect values restored",
			walRecords: []*wal.LogRecord{
				{LSN: wal.LSN{SessionID: 1, SeqID: 1}, CommandID: querylang.CommandSet, Arguments: []string{"key1", "foo"}},
				{LSN: wal.LSN{SessionID: 1, SeqID: 2}, CommandID: querylang.CommandSet, Arguments: []string{"key2", "bar"}},
				{LSN: wal.LSN{SessionID: 1, SeqID: 3}, CommandID: querylang.CommandDel, Arguments: []string{"key1"}},
			},
			wantData: map[string]Data{
				"key1": {err: storage.ErrNotFound},
				"key2": {value: "bar"},
			},
			wantRecords: []*wal.LogRecord{
				{LSN: wal.LSN{SessionID: 1, SeqID: 1}, CommandID: querylang.CommandSet, Arguments: []string{"key1", "foo"}},
				{LSN: wal.LSN{SessionID: 1, SeqID: 2}, CommandID: querylang.CommandSet, Arguments: []string{"key2", "bar"}},
				{LSN: wal.LSN{SessionID: 1, SeqID: 3}, CommandID: querylang.CommandDel, Arguments: []string{"key1"}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
			fs := afero.NewMemMapFs()
			mapStorage := inmemory.NewMapStorage()
			writeRecords(t, fs, "wal.log", test.walRecords)
			controller, err := wal.NewController(
				storage.NewController(mapStorage),
				fs,
				logger,
				10,
				10*time.Millisecond,
				10_000,
				walDirectory,
			)
			require.NoError(t, err)

			ctx, stop := context.WithCancel(context.Background())
			defer stop()
			waiter := sync.WaitGroup{}
			waiter.Add(2)
			go func() {
				defer waiter.Done()
				controller.Serve(ctx)
			}()
			go func() {
				defer waiter.Done()
				for i, command := range test.applyCommands {
					_, err := controller.Execute(command)
					require.NoError(t, err, "command %d: %s", i, command.ID())
				}
				stop()
			}()
			waiter.Wait()

			for key, want := range test.wantData {
				value, err := mapStorage.Get(key)
				if want.err != nil {
					assert.ErrorIs(t, err, want.err, "key %q", key)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, want.value, value)
				}
			}
			gotRecords := readRecords(t, fs)
			require.Equal(t, len(test.wantRecords), len(gotRecords))
			for i, wantRecord := range test.wantRecords {
				assert.Greater(t, gotRecords[i].LSN.SessionID, uint64(0), "record %d", i)
				assert.Equal(t, wantRecord.LSN.SeqID, gotRecords[i].LSN.SeqID, "record %d", i)
				assert.Equal(t, wantRecord.CommandID, gotRecords[i].CommandID, "record %d", i)
				assert.Equal(t, wantRecord.Arguments, gotRecords[i].Arguments, "record %d", i)
			}
		})
	}
}

func writeRecords(tb testing.TB, fs afero.Fs, filename string, records []*wal.LogRecord) {
	tb.Helper()
	if len(records) == 0 {
		return
	}

	file, err := fs.Create(walDirectory + "/" + filename)
	require.NoError(tb, err, "create WAL file")
	defer file.Close()
	encoder := gob.NewEncoder(file)
	require.NoError(tb, encoder.Encode(records), "write WAL file")
}

func readRecords(tb testing.TB, fs afero.Fs) []*wal.LogRecord {
	tb.Helper()

	files, err := afero.ReadDir(fs, walDirectory)
	require.NoError(tb, err, "read WAL directory")

	var records []*wal.LogRecord

	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			continue
		}

		file, err := fs.Open(walDirectory + "/" + fileInfo.Name())
		require.NoError(tb, err, "read WAL file")

		data, err := io.ReadAll(file)
		require.NoError(tb, err, "read WAL file contents")

		buffer := bytes.NewBuffer(data)
		for buffer.Len() > 0 {
			var batch []*wal.LogRecord
			decoder := gob.NewDecoder(buffer)
			err := decoder.Decode(&batch)
			require.NoError(tb, err, "parse WAL records")

			records = append(records, batch...)
		}

		file.Close()
	}

	return records
}
