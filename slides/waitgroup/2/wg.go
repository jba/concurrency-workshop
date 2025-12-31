package wg

import "sync"

// heading WaitGroup with a mutex

// note
// We can synchronize with a mutex.
// !note

// code
type WaitGroup struct {
	// em
	mu sync.Mutex
	// !em
	count int // number of active goroutines
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
	g.count += n
}

// !em

func (g *WaitGroup) Wait() {
	// Wait for g.count to reach 0.
}

// !code
