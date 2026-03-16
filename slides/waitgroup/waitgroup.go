// title Concurrency Patterns
// subtitle
// Demystifying Concurrency

// GopherCon Europe 2026
// !subtitle

package wg

import (
	"sync"
	"sync/atomic"
	"time"
)

////////////////////////////////////
// heading Implementing WaitGroup

// html Unsynchronized version
// cols
// code
type WaitGroup struct {
	count int // number of active goroutines
}

func (g *WaitGroup) Go(f func()) {
	g.Add(1)
	go func() {
		defer g.Done()
		f()
	}()
}

func (g *WaitGroup) Add(n int) { g.count += n }

func (g *WaitGroup) Done() { g.Add(-1) }

func (g *WaitGroup) Wait() {
	// Wait for g.count to reach 0.
}

// !code

// nextcol

// html <div style="height: 5em"></div>

// text
// See [this CL](https://go-review.git.corp.google.com/c/go/+/717760)
// for a recent, subtle change to `WaitGroup.Go`.
// !text
// html <div style="height: 5em"></div>

// !cols

////////////////////////////////////
// heading WaitGroup with a mutex

// code
type WaitGroup_1 struct {
	mu    sync.Mutex // em
	count int        // number of active goroutines
}

func (g *WaitGroup_1) Add(n int) {
	// em
	g.mu.Lock()
	defer g.mu.Unlock()
	// !em
	g.count += n
}

// !code

////////////////////////////////////
// heading WaitGroup with atomics

// /cols
// code
type WaitGroup_2 struct {
	count atomic.Int64 // number of active goroutines // em
}

func (g *WaitGroup_2) Add(n int) {
	g.count.Add(int64(n)) // em
}

// !code

/* text
Atomics work well here, for now

Stdlib implementation uses them
*/

// !cols

////////////////////////////////////
// heading The Wait method

// text
// `Wait` should block until the count is zero.
// It may be called more than once, perhaps concurrently.
// !text

// cols
// code
type WaitGroup_3 struct {
	count atomic.Int64 // number of active goroutines
}

func (g *WaitGroup_3) Wait() {
	for g.count.Load() > 0 {
		time.Sleep(time.Millisecond)
	}
}

// !code
// question
// What's wrong with busy-waiting?
// answer
// - Sleep too long: waste time
// - Sleep too short: waste CPU
// !question

// nextcol

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
// <div style="font-size: 70%; line-height: 1.0">
// (Technically, this is disallowed:
// "Note that calls with a positive delta that occur when the counter is zero must happen before a Wait.")
// </div>
// !question
// !cols

////////////////////////////////////
// heading Fixing busy-waiting Wait

// text
// - Use a channel
// - Close it to broadcast
// !text

// note
// We'll use a channel. Channels are the only way we've seen
// for threads to wait for an event to occur.
//
// The trick here is to close the channel, thereby signaling all readers.
// If we sent something to the channel, that would only wake up one reader.
// !note

// /cols
// code
type WaitGroup_4 struct {
	count atomic.Int64 // number of active goroutines
	// em
	done chan struct{} // closed when count reaches zero
	// !em
}

// em
func NewWaitGroup_4() *WaitGroup_4 {
	return &WaitGroup_4{done: make(chan struct{})}
}

// !em
func (g *WaitGroup_4) Go(f func()) {
	g.add(1)
	go func() {
		defer g.add(-1)
		f()
	}()
}

// !code
// code
func (g *WaitGroup_4) add(n int) {
	c := g.count.Add(int64(n))
	// em
	if c == 0 {
		close(g.done)
	}
	// !em
}

func (g *WaitGroup_4) Wait() {
	// In-class exercise: what goes here?
}

// !code

// !cols

// question
// What should the body of Wait be?
// answer
// `<-g.done`
// That's it!
// !question

// question
// Find the bug in `add`.
// answer
// Two goroutines may both end up on `if c == 0` with c == 0.
// (How?)
// Closing a channel for a second time panics.
// !question

////////////////////////////////////
// heading Back to a mutex

// text
// Need a mutex to perform more than one operation atomically
// !text

// /cols
// code
type WaitGroup_5 struct {
	mu    sync.Mutex
	count int           // number of active goroutines
	done  chan struct{} // closed when count reaches zero
}

func (g *WaitGroup_5) Go(f func()) {
	g.add(1)
	go func() {
		defer g.add(-1)
		f()
	}()
}

// !code
// code
// em
func (g *WaitGroup_5) add(n int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.done == nil {
		g.done = make(chan struct{})
	}
	g.count += n
	if g.count == 0 {
		close(g.done)
		g.done = nil // don't close channel twice
	}
}

// !em
func (g *WaitGroup_5) Wait() {
	// wait for close
	<-g.done
}

// !code
// !cols

// question
// Find the race condition.
// answer
// Wait reads `g.done` without the lock.
// !question

// question
// This WaitGroup can't be used for a second wave of `Go` calls: once the first
// group of goroutines completes, the channel is nil, and `Wait` will never return.
//
// How can we fix that?
// answer
// TODO
// !question

////////////////////////////////////
// heading Fixing the race

// text
// - Channel _operations_ are concurrency-safe
// - But _accessing a variable_ (even one holding a channel) is not
// !text

// /cols
// code
type WaitGroup_6 struct {
	mu    sync.Mutex
	count int           // number of active goroutines
	done  chan struct{} // closed when count reaches zero
}

func (g *WaitGroup_6) Go(f func()) {
	g.add(1)
	go func() {
		defer g.add(-1)
		f()
	}()
}

// !code
// code
func (g *WaitGroup_6) add(n int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.done == nil {
		g.done = make(chan struct{})
	}
	g.count += n
	if g.count == 0 {
		close(g.done)
		g.done = nil // Make sure we only close this channel once.
	}
}

// em
func (g *WaitGroup_6) Wait() {
	// Wait for something to be written to the channel, or for it to be closed.
	g.mu.Lock()
	defer g.mu.Unlock()
	<-g.done
}

// !em
// !code
// !cols

// question
// Find the bug.
// answer
// `Wait` holds the mutex while it's waiting, so `add` can't run to decrement the counter.
// !question

////////////////////////////////////
// heading Fix to previous

// text
// - Channel _operations_ are concurrency-safe
// - But _accessing a variable_ (even one holding a channel) is not
// !text

// code
type WaitGroup_7 struct {
	mu    sync.Mutex
	count int           // number of active goroutines
	done  chan struct{} // closed when count reaches zero
}

func (g *WaitGroup_7) Go(f func()) {
	g.add(1)
	go func() {
		defer g.add(-1)
		f()
	}()
}

func (g *WaitGroup_7) add(n int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.done == nil {
		g.done = make(chan struct{})
	}
	g.count += n
	if g.count == 0 {
		close(g.done)
		g.done = nil // Make sure we only close this channel once.
	}
}

func (g *WaitGroup_7) Wait() {
	// Wait for something to be written to the channel, or for it to be closed.
	// em
	g.mu.Lock()
	d := g.done
	g.mu.Unlock()
	// !em
	<-d
}

// !code
////////////////////////////////////
// heading The real thing

// text
// Actual sync.WaitGroup implementation
// !text
