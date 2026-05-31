package waitgroup

import "sync"

type WaitGroup struct {
	done  chan struct{} // closed when count == 0
	mu    sync.Mutex
	count int // number of active goroutines
}

func NewWaitGroup() *WaitGroup {
	return &WaitGroup{done: make(chan struct{})}
}

func (g *WaitGroup) Add(n int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.count += n
	if g.count == 0 {
		close(g.done)
	}
}

func (g *WaitGroup) Done() {
	g.Add(-1)
}

func (g *WaitGroup) Wait() {}
