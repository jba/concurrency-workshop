package wg

import (
	"sync/atomic"
)

// heading Let's use atomics!

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

func (g *WaitGroup) Add(n int) {
	// em
	g.count.Add(int64(n))
	// !em
}

func (g *WaitGroup) Done() {
	// em
	g.count.Add(-1)
	// !em
}

// !code
