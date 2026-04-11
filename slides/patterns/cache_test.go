package patterns

import (
	"sync/atomic"
	"testing"
)

func TestFib(t *testing.T) {
	for _, test := range []struct {
		in, out int
	}{
		{0, 0},
		{1, 1},
		{2, 1},
		{3, 2},
		{4, 3},
		{5, 5},
	} {
		if g, w := fib(test.in), test.out; g != w {
			t.Errorf("fib(%d) = %d, want %d", test.in, g, w)
		}
	}
}

func TestCache(t *testing.T) {
	calls := 0
	f := func(x int) int {
		calls++
		return x
	}
	m := NewMemo(f)

	check := func(x int) {
		if g := m.Call(x); g != x {
			t.Errorf("got %d, want %d", g, x)
		}
	}

	check(1)
	check(2)
	check(1)

	if calls != 2 {
		t.Errorf("got %d calls, want 2", calls)
	}
}

func TestMemo_3(t *testing.T) {
	calls := 0
	f := func(x int) int {
		calls++
		return x * 2
	}
	m := NewMemo_3(f)

	for i := 0; i < 3; i++ {
		if got := m.Call(10); got != 20 {
			t.Errorf("call %d: got %d, want 20", i, got)
		}
	}
	if calls != 1 {
		t.Errorf("got %d calls, want 1", calls)
	}
}

func TestMemo_3Concurrent(t *testing.T) {
	var calls atomic.Int64
	start := make(chan struct{})
	f := func(x int) int {
		calls.Add(1)
		<-start // make sure multiple goroutines are waiting
		return x * x
	}
	m := NewMemo_3(f)

	const n = 10
	results := make(chan int, n)
	for range n {
		go func() {
			results <- m.Call(5)
		}()
	}

	// Release the function execution
	close(start)

	for range n {
		if got := <-results; got != 25 {
			t.Errorf("got %d, want 25", got)
		}
	}

	if got := calls.Load(); got != 1 {
		t.Errorf("got %d calls, want 1", got)
	}
}

func TestMemo_3Recursive(t *testing.T) {
	// Make sure it works with a recursive function.
	m := NewMemo_3(fib)
	c := make(chan int, 10)
	for range cap(c) {
		go func() {
			c <- m.Call(6)
		}()
	}

	want := fib(6)
	for range cap(c) {
		if g := <-c; g != want {
			t.Errorf("got %d, want %d", g, want)
		}
	}
}
