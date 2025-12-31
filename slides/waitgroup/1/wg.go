package wg

// heading Implementing WaitGroup

// note
// Let's try to implement `sync.WaitGroup` ourselves.
// It has two methods: `Go` and `Wait`.
// We'll start with `Go`.

// All we need to support it is a simple counter, holding
// the number of active goroutines.
// !note

// html <div class='flex'>
// code
type WaitGroup struct {
	count int // number of active goroutines
}

func (g *WaitGroup) Go(f func()) {
	g.add(1)
	go func() {
		defer g.add(-1)
		f()
	}()
}

func (g *WaitGroup) add(n int) {
	g.count += n
}

func (g *WaitGroup) Wait() {
	// Wait for g.count to reach 0.
}

// !code
// text
// - Lots of interesting stuff here.
// - More than you realize.
// !text
// html </div>

// question
// Thoughts?
// answer
// The problem is that there is no synchronization.
// `Go` should be goroutine-safe.
// !question
