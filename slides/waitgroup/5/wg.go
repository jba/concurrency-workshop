package wg

import "sync/atomic"

// heading Fixing busy-waiting Wait

// note
// We'll use a channel. Channels are the only way we've seen
// for threads to wait for an event to occur.
//
// The trick here is to close the channel, thereby signaling all readers.
// If we sent something to the channel, that would only wake up one reader.
// !note

type WaitGroup struct {
	count atomic.Int64  // number of active goroutines
	done  chan struct{} // closed when count reaches zero
}

func NewWaitGroup() *WaitGroup {
	return &WaitGroup{done: make(chan struct{})}
}

func (g *WaitGroup) Go(f func()) {
	g.add(1)
	go func() {
		defer g.add(-1)
		f()
	}()
}

func (g *WaitGroup) add(n int) {
	c := g.count.Add(int64(n))
	if c == 0 {
		close(g.done)
	}
}

// code
func (g *WaitGroup) Wait() {
	// In-class exercise: what goes here?
}

// !code

// question
// What should the body of Wait be?
// answer
// `<-g.done`
// That's it!
// !question

// question
// There is a subtle bug in `add`. Find it.
// answer
// Two goroutines may both end up on `if c == 0` with c == 0.
// (How?)
// Closing a channel for a second time panics.
// !question
