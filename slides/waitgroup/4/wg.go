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

// div.flex
// code
type WaitGroup struct {
	count atomic.Int64 // number of active goroutines
}

func (g *WaitGroup) Go(f func()) {
	g.Add(1)
	go func() {
		defer g.Add(-1)
		f()
	}()
}

func (g *WaitGroup) Add(n int) {
	g.count.Add(int64(n))
}

func (g *WaitGroup) Done() { g.Add(-1) }

// em
func (g *WaitGroup) Wait() {
	for g.count.Load() > 0 {
		time.Sleep(time.Millisecond)
	}
}

// !em
// !code

// html <div> <!-- one child for flex -->
// question

// What's wrong with busy-waiting?
// answer
// - Sleep too long: waste time
// - Sleep too short: waste CPU
// !question

// question
// Find the bug.
// answer
// `Wait` might not notice a 0 count.
//
// <div class="interleave" style="font-size: 70%">

// | G1 | G2 |
// | -- | -- |
// |   | Add(1) |
// | Load | |
// | Sleep | |
// |     | Done() |
// |    | Add(1) |
// | Load | |
// | Sleep | |

// </div>
//
//
// <div style="font-size: 50%">
// (Technically, this is disallowed:
// "Note that calls with a positive delta that occur when the counter is zero must happen before a Wait.")
// </div>
// !question
// html </div>
// !div.flex
