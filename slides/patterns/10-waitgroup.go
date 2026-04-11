package patterns

import (
	"sync"
	"time"
)

// title Implementing WaitGroup

////////////////////////////////////
// heading WaitGroup unsynchronized

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
func (g *WaitGroup_1) Done() { g.Add(-1) }

func (g *WaitGroup_1) Wait() {
	// Wait for g.count to reach 0.
}

// !code

////////////////////////////////////
// heading Busy-waiting

// text
// `Wait` should block until the count is zero.
// !text

// cols
// code
type WaitGroup_2 struct {
	mu    sync.Mutex
	count int
}

func (g *WaitGroup_2) Add(n int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.count += n
}

func (g *WaitGroup_2) Done() { g.Add(-1) }

// !code
// nextcol

// code bad
func (g *WaitGroup_2) Wait() {
	g.mu.Lock()
	defer g.mu.Unlock()
	for g.count > 0 {
		time.Sleep(time.Millisecond)
	}
}

// !code

// question What's the problem here?
// answer
// `Wait` holds the mutex for its entire lifetime,
// so `Done` is blocked.
// !question
// !cols

// //////////////////////////////////
// heading Busy-waiting, 2
// cols
type WaitGroup_3 struct {
	mu    sync.Mutex
	count int
}

// code weak

func (g *WaitGroup_3) Wait() {
	for {
		g.mu.Lock()
		c := g.count
		g.mu.Unlock()
		if c <= 0 {
			break
		}
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
// Find the issue.
// answer
// `Wait` might not notice a 0 count.
//
// <div class="interleave" style="font-size: 70%">

// | G-wait | G-other |
// | -- | -- |
// |   | Add(1) |
// | c := g.count| |
// | Sleep | |
// |     | Done() |
// |    | Add(1)\* |
// | c := g.count| |
// | Sleep | |

// </div>
//
//
// <div style="font-size: 70%; line-height: 1.0; margin-top: 20px">
// *Technically, this is disallowed:
// "Note that calls with a positive delta that occur when the counter is zero must happen before a Wait."
// </div>
// !question
// !cols

////////////////////////////////////
// heading Fixing busy-waiting Wait

// text `Wait` should actually wait.
//
// question
// What synchronization feature will make a goroutine
// wait until something happens?<br/>
// And how should we use it?
// answer
// A channel.
//
// We should close it when `count` is zero, to broadcast to
// all waiting goroutines.
// code
type WaitGroup_answer struct {
	done  chan struct{} // closed when count == 0
	mu    sync.Mutex
	count int // number of active goroutines
}

// !code
// !question

////////////////////////////////////
// heading Wait without busy-waiting

// cols
// code
type WaitGroup_5 struct {
	done  chan struct{} // closed when count == 0
	mu    sync.Mutex
	count int // number of active goroutines
}

func NewWaitGroup() *WaitGroup_5 {
	return &WaitGroup_5{done: make(chan struct{})}
}

func (g *WaitGroup_5) Add(n int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.count += n
	if g.count == 0 {
		close(g.done)
	}
}

// !code

// nextcol

// question
// What should the body of `Wait` be?
// answer
// code
func (g *WaitGroup_5) Wait() {
	<-g.done
}

// !code
// !question

// question What happens if there is another "round" with the same `WaitGroup`?
// answer
// `Wait` immediately returns, and `Add` panics when the channel is closed again.
// !question

// !cols

// heading Exercise: Improvements

// text
// 1. Get rid of the the constructor, so a zero `WaitGroup` is ready to use.
// 2. Allow multiple "rounds": after `Wait` returns, the `WaitGroup` can
// be used again by calling `Add` multiple times, then `Wait`.
//
// Hint: modify `g.done`.
//
// Not that channel _operations_ are concurrency-safe,
// but reading and writing a channel _variable_ is not.
// !text

// heading Exercise solution

// cols
// code
// include ../../exercises/waitgroup/solution/waitgroup.go /^type/ /^\}$/
// include ../../exercises/waitgroup/solution/waitgroup.go /^func.*Add\(/ /^\}$/
// !code

// nextcol
// code
// include ../../exercises/waitgroup/solution/waitgroup.go /^func.*Wait\(/ /^\}$/
// !code
// !cols
