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
	g.Add(1)
	go func() {
		defer g.Done()
		f()
	}()
}

// em
func (g *WaitGroup) Add(n int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.count += n
}

// !em

func (g *WaitGroup) Done() { g.Add(-1) }

func (g *WaitGroup) Wait() {
	// Wait for g.count to reach 0.
}

// !code
