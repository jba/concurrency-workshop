package hedging

import (
	"context"
	"testing"
	"time"
)

// getResult calls both method1 and method2
// concurrently, returning the first result it gets.
// Before returning, it should cancel the other computation.
func getResult(ctx context.Context, input int) int {
	// TODO
	return 0
}

// NOTE: In a real hedging situation, both methods would return the same (or nearly the
// same) results. For testing, our methods return very different results.

// method1 returns input.
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

// method1 returns twice input.
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
