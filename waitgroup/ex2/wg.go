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
	count atomic.Int64 // number of active goroutines
}

func (g *WaitGroup) Add(n int) {
	g.count.Add(int64(n))
}

func (g *WaitGroup) Done() {
	g.count.Add(-1)
}

// !code

// func (g *WaitGroup) Wait() {
// 	g.mu.Lock()
// 	defer g.mu.Unlock()
// 	for g.count > 0 {
// 		time.Sleep(time.Millisecond)
// 	}
// }

// locking during Wait deadlocks if anyone adds something.
