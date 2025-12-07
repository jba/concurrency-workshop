package wg

import "time"

// heading Implementing WaitGroup: Add and Done

// note
// Let's try to implement sync.WaitGroup ourselves.
// It has three methods: Add, Done and Wait.
// We'll start with Add and Done.
// All we need to support them is a simple counter, holding
// the number of started but not finished goroutines.

// What do you think about this solution?

// Problem: no synchronization.
// !note

// code
type WaitGroup struct {
	count int // number of active goroutines
}

func (g *WaitGroup) Add(n int) {
	g.count += n
}

func (g *WaitGroup) Done() {
	g.count--
}

// !code

func (g *WaitGroup) Wait() {
	for g.count > 0 {
		time.Sleep(time.Millisecond)
	}
}
