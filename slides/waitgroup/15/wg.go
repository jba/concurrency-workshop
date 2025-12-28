package wg

import (
	"sync"
	"time"
)

// heading With a mutex

// note
// We can synchronize with a mutext.
// !note

// code
type WaitGroup struct {
	mu    sync.Mutex
	count int // number of active goroutines
}

func (g *WaitGroup) Go(f func()) {
	g.mu.Lock()
	g.count++
	g.mu.Unlock()
	go func() {
		defer func() {
			g.mu.Lock()
			g.count--
			g.mu.Unlock()
		}()
		f()
	}()
}

func (g *WaitGroup) Wait() {
	// Wait for g.count to reach 0.
}

// !code

// question
// Find the bug.
// answer
// If we lock for all of Go, there can be only one goroutine
// `Go` should be goroutine-safe.
// !question

// func (g *WaitGroup) Wait() {
// 	for g.count > 0 {
// 		time.Sleep(time.Millisecond)
// 	}
// }
