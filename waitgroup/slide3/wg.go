package wg

import (
	"sync/atomic"
)

// heading Atomic bomb

// note
// We should check that `Done` isn't called when the count is zero.
// Let's add that check.
// !note

// code
type WaitGroup struct {
	count atomic.Int64 // number of active goroutines
}

func (g *WaitGroup) Add(n int) {
	g.count.Add(int64(n))
}

func (g *WaitGroup) Done() {
	// em
	if g.count.Load() <= 0 {
		panic("WaitGroup.Done called without a matching Add")
	}
	// !em
	g.count.Add(-1)
}

// !code

// question
// What do you think about this code?
// Find the bug (if any).
// answer
// Uh-oh! we have a TOCTOU race (Time Of Check-Time Of Use).

// Explain how (that is, provide an interleaving where)
// `g.count` can become negative.

// That's a pitfall of using atomics: when you need to make the code more
// complicated, you may be tempted to make the smallest change.

// !question
