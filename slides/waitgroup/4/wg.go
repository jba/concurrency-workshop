package wg

import (
	"sync/atomic"
	"time"
)

// heading The Wait method

// note
// Let's turn our attention to the `Wait` method.
// It should block until the count is zero.

// `Wait` may be called more than once, perhaps concurrently.

// Here is one possible implementation.
// !note

// code
type WaitGroup struct {
	count atomic.Int64 // number of active goroutines
}

func (g *WaitGroup) Go(f func()) {
	g.add(1)
	go func() {
		defer g.add(-1)
		f()
	}()
}

func (g *WaitGroup) add(n int) {
	g.count.Add(int64(n))
}

// em
func (g *WaitGroup) Wait() {
	for g.count.Load() > 0 {
		time.Sleep(time.Millisecond)
	}
}

// !em

// !code

// question
// What's wrong with busy-waiting?
// answer
// There's no perfect time to sleep. You may sleep too long, wasting time,
// or too short, wasting CPU.
// !question

// question
// Find the bug (if any).
// answer
// The count can go to 0, then back up, and Wait won't notice.
// !question
//
