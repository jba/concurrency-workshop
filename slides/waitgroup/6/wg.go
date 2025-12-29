package wg

import "sync"

// heading Back to a mutex

// text
// Need a mutex to perform more than one operation atomically
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

// em
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

// !em

func (g *WaitGroup) Wait() {
	// Wait for something to be written to the channel, or for it to be closed.
	<-g.done
}

// question
// Find the bug.
// answer
// Wait reads `g.done` without the lock.
// !question

// question
// This WaitGroup can't be used for a second wave of `Go` calls: once the first
// group of goroutines completes, the channel is nil, and `Wait` will never return
// (https://go.dev/ref/spec#Receive_operator:
// "Receiving from a nil channel blocks forever.")
//
// How can we fix that?
// answer
// TODO
// !question
