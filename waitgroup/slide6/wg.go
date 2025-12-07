package wg

import (
	"sync"
	"time"
)

// heading Fixing busy-waiting Wait

// note
// Here is a fix: hold the lock only to get the count.
// !note

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

// code
func (g *WaitGroup) Wait() {
	for {
		g.mu.Lock()
		c := g.count
		g.mu.Unlock()
		if c <= 0 {
			break
		}
		time.Sleep(time.Millisecond)
	}
}

// !code

// note
// It's fine to omit the `defer` here. The locked section is tiny.
// !note

// question
// This is busy-waiting. Why is it bad?
// answer
// There's no perfect time to sleep. You may sleep too long, wasting time,
// or too short, wasting CPU.
// !question
