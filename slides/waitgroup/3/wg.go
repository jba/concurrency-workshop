package wg

import (
	"sync/atomic"
)

// heading WaitGroup with atomics

// div.flex
// code
type WaitGroup struct {
	// em
	count atomic.Int64 // number of active goroutines
	// !em
}

func (g *WaitGroup) Go(f func()) {
	g.Add(1)
	go func() {
		defer g.Add(-1)
		f()
	}()
}

// em
func (g *WaitGroup) Add(n int) {
	g.count.Add(int64(n))
}

// !em

func (g *WaitGroup) Done() { g.Add(-1) }

func (g *WaitGroup) Wait() {
	// Wait for g.count to reach 0.
}

// !code
/* text
- Atomics work well here, for now

- Stdlib implementation uses them
*/

// !div.flex
