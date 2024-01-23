package engine

import "sync/atomic"

type IDGenerator struct {
	current atomic.Uint64
}

func (g *IDGenerator) NextSeqID() uint64 {
	for {
		old := g.current.Load()
		if g.current.CompareAndSwap(old, old+1) {
			return old + 1
		}
	}
}
