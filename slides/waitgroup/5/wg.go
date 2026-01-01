package wg

import "sync/atomic"

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

// div.flex
// code
type WaitGroup struct {
	count atomic.Int64 // number of active goroutines
	// em
	done chan struct{} // closed when count reaches zero
	// !em
}

// em
func NewWaitGroup() *WaitGroup {
	return &WaitGroup{done: make(chan struct{})}
}

// !em
func (g *WaitGroup) Go(f func()) {
	g.add(1)
	go func() {
		defer g.add(-1)
		f()
	}()
}

// !code
// code
func (g *WaitGroup) add(n int) {
	c := g.count.Add(int64(n))
	// em
	if c == 0 {
		close(g.done)
	}
	// !em
}

func (g *WaitGroup) Wait() {
	// In-class exercise: what goes here?
}

// !code

// !div.flex

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
