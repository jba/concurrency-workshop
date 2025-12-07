package wg

import (
	"sync"
	"time"
)

// heading WaitGroup with Mutex

// note
// Let's use a mutex.
// !note

// code
type WaitGroup struct {
	mu    sync.Mutex
	count int // number of active goroutines
}

func (g *WaitGroup) Add(n int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.count += n
}

func (g *WaitGroup) Done() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.count <= 0 {
		panic("WaitGroup.Done called without a matching Add")
	}
	g.count--
}

// !code

func (g *WaitGroup) Wait() {
	g.mu.Lock()
	defer g.mu.Unlock()
	for g.count > 0 {
		time.Sleep(time.Millisecond)
	}
}

// locking during Wait deadlocks if anyone adds something.
