package wg

// heading Implementing WaitGroup

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
	g.Add(1)
	go func() {
		defer g.Done()
		f()
	}()
}

func (g *WaitGroup) Add(n int) {
	g.count += n
}

func (g *WaitGroup) Done() { g.Add(-1) }

func (g *WaitGroup) Wait() {
	// Wait for g.count to reach 0.
}

// !code

// question
// Thoughts?
// answer
// The problem is that there is no synchronization.
// `Go` should be goroutine-safe.
// !question
