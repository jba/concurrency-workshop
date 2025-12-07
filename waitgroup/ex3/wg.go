package wg

import (
	"sync/atomic"
)

// heading Atomic bomb

// note
// We should check that Done isn't called when the count is zero.
// Let's add that check.
// WDYT about this code?
// !note

// code
type WaitGroup struct {
	count atomic.Int64 // number of active goroutines
}

func (g *WaitGroup) Add(n int) {
	g.count.Add(int64(n))
}

func (g *WaitGroup) Done() {
	if g.count.Load() <= 0 {
		panic("WaitGroup.Done called without a matching Add")
	}
	g.count.Add(-1)
}

// !code

// note
// Uh-oh! we have a TOCTOU race (Time Of Check-Time Of Use).
// That's a pitfall of using atomics.
// !note
