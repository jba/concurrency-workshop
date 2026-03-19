package waitgroup

import (
	"sync/atomic"
	"testing"
)

func TestWaitGroup(t *testing.T) {
	var wg WaitGroup
	var count atomic.Int64

	for range 100 {
		wg.Go(func() {
			count.Add(1)
		})
	}
	wg.Wait()

	if got := count.Load(); got != 100 {
		t.Errorf("count = %d, want 100", got)
	}
}

func TestWaitGroupZeroValue(t *testing.T) {
	// A zero WaitGroup should be ready to use
	var wg WaitGroup
	wg.Wait() // Should not block when no goroutines added
}

func TestWaitGroupMultipleRounds(t *testing.T) {
	var wg WaitGroup
	var count atomic.Int64

	// First round
	for range 50 {
		wg.Go(func() {
			count.Add(1)
		})
	}
	wg.Wait()

	if got := count.Load(); got != 50 {
		t.Errorf("after round 1: count = %d, want 50", got)
	}

	// Second round
	for range 50 {
		wg.Go(func() {
			count.Add(1)
		})
	}
	wg.Wait()

	if got := count.Load(); got != 100 {
		t.Errorf("after round 2: count = %d, want 100", got)
	}
}
