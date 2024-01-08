package network_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/strider2038/key-value-database/internal/database/network"
)

func TestSemaphore(t *testing.T) {
	const ticketsCount = 5
	const goroutinesCount = 20

	semaphore := network.NewSemaphore(ticketsCount)
	n := atomic.Int32{}

	wg := sync.WaitGroup{}
	wg.Add(goroutinesCount)
	for i := 0; i < goroutinesCount; i++ {
		go func() {
			defer wg.Done()

			semaphore.Acquire()
			n.Add(1)
			nn := n.Load()
			t.Log("active processes: ", nn)
			if nn < 1 || nn > ticketsCount {
				t.Error("unexpected active processes count: ", nn)
			}
			n.Add(-1)
			semaphore.Release()
		}()
	}
	wg.Wait()
}
