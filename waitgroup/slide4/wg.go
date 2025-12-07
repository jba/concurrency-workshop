package wg

import (
	"sync"
	"time"
)

// heading WaitGroup with mutex

// note
// We better switch to a mutex.

// It's marginally clumsier, but if you defer the unlock,
// you're probably fine.

// It's a bit slower, but you probably won't care.

// And it's much safer!
// !note

// code
type WaitGroup struct {
	// em
	mu sync.Mutex
	// !em
	count int // number of active goroutines
}

func (g *WaitGroup) Add(n int) {
	// em
	g.mu.Lock()
	defer g.mu.Unlock()
	// !em
	g.count += n
}

func (g *WaitGroup) Done() {
	// em
	g.mu.Lock()
	defer g.mu.Unlock()
	// !em
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
