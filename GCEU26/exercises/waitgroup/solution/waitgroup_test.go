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

func TestWaitGroupReuseConcurrent(t *testing.T) {
	var wg WaitGroup
	for round := range 10 {
		var count atomic.Int64
		const n = 100
		const waiters = 10

		var waitersWg WaitGroup
		waitersWg.Add(waiters)
		for range waiters {
			go func() {
				defer waitersWg.Done()
				wg.Wait()
			}()
		}

		// Add some delay to make sure waiters have a chance to start
		// (though wg.Wait() will return immediately if count is 0,
		// which is what we want to test too)

		for range n {
			wg.Go(func() {
				count.Add(1)
			})
		}
		wg.Wait()
		waitersWg.Wait()

		if got := count.Load(); got != n {
			t.Errorf("round %d: count = %d, want %d", round, got, n)
		}
	}
}
