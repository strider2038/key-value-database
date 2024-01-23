package engine_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/strider2038/key-value-database/internal/database/engine"
)

func TestIDGenerator_NextSeqID(t *testing.T) {
	ids := make(map[uint64]struct{})
	generator := engine.IDGenerator{}
	const count = 100

	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			id := generator.NextSeqID()

			mu.Lock()
			ids[id] = struct{}{}
			mu.Unlock()
		}()
	}
	wg.Wait()

	assert.Len(t, ids, count)
	for i := 1; i <= count; i++ {
		assert.Contains(t, ids, uint64(i))
	}
}
