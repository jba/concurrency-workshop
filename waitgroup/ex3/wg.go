package wg

import (
	"sync/atomic"
)

// note
// We should check that Done isn't called when the count is zero.
// Let's add that check.
// WDYT about this code?

// Uh-oh! we have a TOCTOU race (Time Of Check-Time Of Use).
// That's a pitfall of using atomics.
// end note

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

// end code

// func (g *WaitGroup) Wait() {
// 	g.mu.Lock()
// 	defer g.mu.Unlock()
// 	for g.count > 0 {
// 		time.Sleep(time.Millisecond)
// 	}
// }

// locking during Wait deadlocks if anyone adds something.
