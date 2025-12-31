package wg

import (
	"sync/atomic"
)

// heading WaitGroup with atomics

/* text
- Atomics work well here, for now

- Trouble with `Wait`

- Stdlib implementation uses them
*/

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
	g.add(1)
	go func() {
		defer g.add(-1)
		f()
	}()
}

// em
func (g *WaitGroup) add(n int) {
	g.count.Add(int64(n))
}

// !em

func (g *WaitGroup) Wait() {
	// Wait for g.count to reach 0.
}

// !code
