package hedging

import (
	"context"
	"testing"
	"time"
)

func getResult(ctx context.Context, input int) int {
	c := make(chan int, 2)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() { c <- method1(ctx, input) }()
	go func() { c <- method2(ctx, input) }()
	return <-c
}

// NOTE: In a real hedging situation, both methods would return the same (or nearly the
// same) results. For testing, our methods return very different results.

// method1 returns its input.
// It does so immediately if input is positive.
// It sleeps first if input is negative.
func method1(ctx context.Context, input int) int {
	select {
	case <-time.After(max(0, -time.Duration(input)*time.Millisecond)):
		return input
	case <-ctx.Done():
		return 0
	}
}

// method1 returns twice its input.
// It does so immediately if input is negative.
// It sleeps first if input is positive.
func method2(ctx context.Context, input int) int {
	select {
	case <-time.After(max(0, time.Duration(input)*time.Millisecond)):
		return 2 * input
	case <-ctx.Done():
		return 0
	}
}

func Test(t *testing.T) {
	ctx := context.Background()
	// Positive input: method2 sleeps, result is input.
	if g, w := getResult(ctx, 20), 20; g != w {
		t.Errorf("got %d, want %d", g, w)
	}
	// Negative input: method1 sleeps, result is double input.
	if g, w := getResult(ctx, -20), -40; g != w {
		t.Errorf("got %d, want %d", g, w)
	}
}
