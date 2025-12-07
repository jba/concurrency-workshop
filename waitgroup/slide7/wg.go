package wg

import (
	"sync"
	"time"
)

// heading Exercise: implementing Wait

// note
// As an exercise, implement Wait correctly.
// Remember, it may be called more than once, and concurrently.

// You can assume that when the count reaches zero, it will
// stay there.

// You can also add a `NewWaitGroup` constructor to simplify initialization.
//
// For an extra challenge, try this exercise without those assumptions.
// !note

// question
// Here's a hint:

// A correct implementation will block the goroutine that calls `Wait`.
// We've only learned about one feature that blocks goroutines.
// What is it?
// answer
// A channel.
// !question

// note
// You're going to want to store one of those in the WaitGroup struct.

// Use it in `Wait`.

// You have to do something in `Done` too.
// !note

type WaitGroup struct {
	mu    sync.Mutex
	count int // number of active goroutines
	done  chan struct{}
}

func NewWaitGroup() *WaitGroup {
	return &WaitGroup{done: make(chan struct{})}
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
	if g.count == 0 {
		close(g.done)
	}
}

func (g *WaitGroup) Wait() {
	<-g.done
}

// note
// You'll find the answer in waitgroup/slide7/wg.go in the workshop repo.
// !note
