package wg

// heading Implementing WaitGroup

// note
// Let's try to implement `sync.WaitGroup` ourselves.
// It has two methods: `Go` and `Wait`.
// We'll start with `Go`.

// All we need to support it is a simple counter, holding
// the number of active goroutines.
// !note

// div.flex
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

// html <div> <!-- one child for flex -->

// html <div style="height: 5em"></div>

// text
// See [this CL](https://go-review.git.corp.google.com/c/go/+/717760)
// for a recent, subtle change to `Go`.
// !text
// html <div style="height: 5em"></div>

// question
// Any thoughts about how we're using `count`?
// answer
// The problem is that there is no synchronization.
// `Go` should be goroutine-safe.
// !question
// html </div>
// !div.flex
