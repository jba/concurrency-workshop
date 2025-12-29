package wg

import "sync"

// heading

// text
// - Channel _operations_ are concurrency-safe
// - But _accessing a variable_ (even one holding a channel) is not
// !text

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

func (g *WaitGroup) Wait() {
	// Wait for something to be written to the channel, or for it to be closed.
	// em
	g.mu.Lock()
	d := g.done
	g.mu.Unlock()
	// !em
	<-d
}
