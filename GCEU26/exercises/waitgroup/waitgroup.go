package waitgroup

import "sync"

type WaitGroup struct {
	mu    sync.Mutex
	done  chan struct{} // closed when count == 0
	count int // number of active goroutines
}

func (g *WaitGroup) Add(n int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.done == nil {
		g.done = make(chan struct{})
	}
	g.count += n
	if g.count == 0 {
		close(g.done)
	}
}

func (g *WaitGroup) Done() {
	g.Add(-1)
}

func (g *WaitGroup) Wait() { <-g.done }
