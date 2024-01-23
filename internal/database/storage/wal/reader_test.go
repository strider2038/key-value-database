package wal_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/strider2038/key-value-database/internal/database/querylang"
	"github.com/strider2038/key-value-database/internal/database/storage/wal"
)

func TestReader_ReadRecords(t *testing.T) {
	fs := afero.NewMemMapFs()
	writeRecords(t, fs, "wal_02_01.log", []*wal.LogRecord{
		{
			LSN:       wal.LSN{SessionID: 2, SeqID: 3},
			CommandID: querylang.CommandSet,
			Arguments: []string{"k", "v"},
		},
		{
			LSN:       wal.LSN{SessionID: 2, SeqID: 1},
			CommandID: querylang.CommandSet,
			Arguments: []string{"k", "v"},
		},
		{
			LSN:       wal.LSN{SessionID: 2, SeqID: 2},
			CommandID: querylang.CommandSet,
			Arguments: []string{"k", "v"},
		},
	})
	writeRecords(t, fs, "wal_01_01.log", []*wal.LogRecord{
		{
			LSN:       wal.LSN{SessionID: 1, SeqID: 2},
			CommandID: querylang.CommandSet,
			Arguments: []string{"k", "v"},
		},
		{
			LSN:       wal.LSN{SessionID: 1, SeqID: 1},
			CommandID: querylang.CommandSet,
			Arguments: []string{"k", "v"},
		},
	})

	records, err := wal.NewReader(fs, walDirectory).ReadRecords()

	require.NoError(t, err)
	wantLSNs := []wal.LSN{
		{SessionID: 1, SeqID: 1},
		{SessionID: 1, SeqID: 2},
		{SessionID: 2, SeqID: 1},
		{SessionID: 2, SeqID: 2},
		{SessionID: 2, SeqID: 3},
	}
	require.Len(t, records, len(wantLSNs))
	for i, wantLSN := range wantLSNs {
		assert.Equal(t, wantLSN, records[i].LSN)
	}
}
