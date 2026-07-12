package waitgroup

import "sync"

// WaitGroup waits for a collection of goroutines to finish.
// A zero WaitGroup is ready to use.
// After Wait returns, the WaitGroup can be reused for another round.
type WaitGroup struct {
	mu    sync.Mutex
	count int
	done  chan struct{}
}

func (g *WaitGroup) Add(n int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	// Lazily create the channel.
	if g.done == nil {
		g.done = make(chan struct{})
	}
	g.count += n
	if g.count == 0 {
		close(g.done)
		g.done = nil // Allow reuse for next round.
	}
}

func (g *WaitGroup) Done() {
	g.Add(-1)
}

func (g *WaitGroup) Go(f func()) {
	g.Add(1)
	go func() {
		defer g.Done()
		f()
	}()
}

func (g *WaitGroup) Wait() {
	g.mu.Lock()
	d := g.done
	g.mu.Unlock()
	if d != nil {
		<-d
	}
}
