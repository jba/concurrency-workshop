package wg

import (
	"sync/atomic"
)

// heading Atomics

// note
// This looks like a good place to use atomics, because we're just
// incrementing and decrementing a counter. So let's do that.
// !note

// code
type WaitGroup struct {
	// em
	count atomic.Int64 // number of active goroutines
	// !em
}

func (g *WaitGroup) Go(f func()) {
	// em
	g.count.Add(1)
	// !em
	go func() {
		// em
		defer g.count.Add(-1)
		// !em
		f()
	}()
}

func (g *WaitGroup) Wait() {
	// Wait for g.count to reach 0.
}

// !code
