package wg

import (
	"sync"
	"time"
)

// heading Back to a mutex

// note
// Atomics were convenient, but they led us down the wrong path.
// We need a mutex here to atomically decrement count and close the channel.
//
// Does that mean we couldn't use atomics?
// No! In fact, the standard library WaitGroup uses atomics.
// But this implementation is much easier to understand.
// !note

// note

type WaitGroup struct {
	mu    sync.Mutex
	count int           // number of active goroutines
	done  chan struct{} // closed when count reaches zero
}

func NewWaitGroup() *WaitGroup {
	return &WaitGroup{done: make(chan struct{})}
}

func (g *WaitGroup) Go(f func()) {
	g.mu.Lock()
	g.count++
	g.mu.Unlock()

	go func() {
		defer func() {
			g.mu.Lock()
			defer g.mu.Unlock()
			g.count--
			if g.count == 0 && g.done != nil {
				close(g.done)
				g.done = nil
			}
		}()

		f()
	}()
}

func (g *WaitGroup) Wait() {
	// Wait for something to be written to the channel, or for it to be closed.
	<-g.done
}

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
