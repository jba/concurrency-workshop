package wg

import "sync"

// heading Fixing the race

// text
// - Channel _operations_ are concurrency-safe
// - But _accessing a variable_ (even one holding a channel) is not
// !text

// div.flex
// code
type WaitGroup struct {
	mu    sync.Mutex
	count int           // number of active goroutines
	done  chan struct{} // closed when count reaches zero
}

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
func (g *WaitGroup) Wait() {
	// Wait for something to be written to the channel, or for it to be closed.
	g.mu.Lock()
	defer g.mu.Unlock()
	<-g.done
}

// !em
// !code
// !div.flex

// question
// Find the bug.
// answer
// `Wait` holds the mutex while it's waiting, so `add` can't run to decrement the counter.
// !question
