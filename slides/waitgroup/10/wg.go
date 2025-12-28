package wg

import "time"

// heading Implementing WaitGroup: Add and Done

// note
// Let's try to implement `sync.WaitGroup` ourselves.
// It has two methods: `Go` and `Wait`.
// We'll start with `Go`.

// All we need to support it is a simple counter, holding
// the number of active goroutines.

// !note

// code
type WaitGroup struct {
	count int // number of active goroutines
}

func (g *WaitGroup) Go(f func()) {
	g.count++
	go func() {
		defer func() { g.count-- }()
		f()
	}()
}

func (g *WaitGroup) Wait() {
	// Wait for g.count to reach 0.
}

// !code

// question
// What do you think about this solution?
// answer
// The problem is that there is no synchronization.
// `Go` should be goroutine-safe.
// !question

// func (g *WaitGroup) Wait() {
// 	for g.count > 0 {
// 		time.Sleep(time.Millisecond)
// 	}
// }
