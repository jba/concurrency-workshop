package wg

import (
	"sync"
	"time"
)

// heading The Wait method

// note
// Let's turn our attention to the `Wait` method.
// It should block until the count is zero.

// We're going to assume throughout the rest of this section
// that once `Wait` is called, `Add` will never be called again.
// The real `WaitGroup` handles that case, but for simplicity, we will not.

// However, `Wait` may be called more than once, perhaps concurrently.

// Here is one possible implementation.
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

// em
func (g *WaitGroup) Wait() {
	g.mu.Lock()
	defer g.mu.Unlock()
	for g.count > 0 {
		time.Sleep(time.Millisecond)
	}
}

// !em

// !code

// question
// What do you think of this?
// Find the bug (if any).
// answer
// Since `Wait` holds the mutex for the entire time it's running, `Done` can
// never run. If `g.count` is already zero, `Wait` exits immediately. But if
// it is positive, `Wait` will run forever.
// !question
