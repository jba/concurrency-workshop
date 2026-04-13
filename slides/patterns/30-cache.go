package patterns

import (
	"sync"
)

// xtitle A concurrent function cache

////////////////////////////////////
// heading Using the cache

// text Memoize a function: cache results of previous calls.

// cols
// code
// fib computes the nth number in
// the Fibonacci sequence.
func fib(n int) int {
	switch n {
	case 0:
		return 0
	case 1:
		return 1
	default:
		return fib(n-1) + fib(n-2)
	}
}

// !code
// nextcol

func ex() {
	// code
	memo := NewMemo(fib)
	memo.Call(3) // fib(3), fib(2), fib(1), fib(0)
	memo.Call(4) // only calls fib(4)

	// !code
}

// !cols

////////////////////////////////////
// heading First version: not concurrency-safe

// cols

// code weak
// Memo memoizes function calls.
// It is NOT safe for use by multiple goroutines.
type Memo[In comparable, Out any] struct {
	f func(In) Out
	m map[In]Out
}

func NewMemo[In comparable, Out any](
	f func(In) Out,
) *Memo[In, Out] {
	return &Memo[In, Out]{f: f, m: map[In]Out{}}
}

// !code
// nextcol
// code weak

func (m *Memo[In, Out]) Call(in In) Out {
	out, ok := m.m[in]
	if !ok {
		out = m.f(in)
		m.m[in] = out
	}
	return out
}

// !code
// !cols
// text (Adapted from Donovan and Kernighan, _The Go Programming Language_)
////////////////////////////////////
// heading Safe but not concurrent

// cols
// code weak
// Memo memoizes function calls.
// It is safe for use by multiple goroutines
// if the function is.
type Memo_2[In comparable, Out any] struct {
	f  func(In) Out
	mu sync.Mutex // em
	m  map[In]Out
}

func NewMemo_2[In comparable, Out any](
	f func(In) Out,
) *Memo_2[In, Out] {
	return &Memo_2[In, Out]{f: f, m: map[In]Out{}}
}

// !code
// nextcol
// code
func (m *Memo_2[In, Out]) Call(in In) Out {
	// em
	m.mu.Lock()
	defer m.mu.Unlock()
	// !em
	out, ok := m.m[in]
	if !ok {
		out = m.f(in)
		m.m[in] = out
	}
	return out
}

// !code

////////////////////////////////////
// heading Safe and concurrent Memo, 1

// code
// Memo memoizes function calls.
// It is safe for use by multiple goroutines,
// if the function is.
type Memo_3[In comparable, Out any] struct {
	f  func(In) Out
	mu sync.Mutex
	m  map[In]*entry[Out] // em
}

func NewMemo_3[In comparable, Out any](
	f func(In) Out,
) *Memo_3[In, Out] {
	return &Memo_3[In, Out]{f: f, m: map[In]*entry[Out]{}}
}

type entry[Out any] struct {
	out   Out
	waitc chan struct{}
}

// !code
////////////////////////////////////
// heading Safe and concurrent Memo, 2

// code
func (m *Memo_3[In, Out]) Call(in In) Out {
	m.mu.Lock()
	e := m.m[in]
	if e == nil {
		// This is the first request for this key.
		// This goroutine is responsible for computing the value.
		e = &entry[Out]{waitc: make(chan struct{})}
		m.m[in] = e
		m.mu.Unlock()
		e.out = m.f(in)
		close(e.waitc) // broadcast readiness to all waiters
	} else {
		// This key is already being computed or is ready.
		m.mu.Unlock()
		<-e.waitc
	}
	return e.out
}

// !code

////////////////////////////////////
// heading Safe and concurrent Memo: locking

// cols

// code
func (m *Memo_3[In, Out]) Call_1(in In) Out {
	m.mu.Lock()
	e := m.m[in]
	if e == nil {
		// This is the first request for this key.
		// This goroutine is responsible for computing the value.
		e = &entry[Out]{waitc: make(chan struct{})}
		m.m[in] = e
		m.mu.Unlock()
		// em
		e.out = m.f(in)
		close(e.waitc) // broadcast readiness to all waiters
		// !em
	} else {
		// This key is already being computed or is ready.
		m.mu.Unlock()
		<-e.waitc // em
	}
	return e.out // em
}

// !code

// text Unlocked lines are emphasized.

// nextcol

// question Why is it safe to access to `e.waitc` unlocked?
// answer
// It's immutable (unchanged after initialization)
// !question

// question Why is it safe to access `e.out` unlocked?
// answer
// - The first goroutine writes `e.out`, then closes the channel.
// - Other goroutines wait for the close, then read `e.out`.
// - The close _happens before_ the wait returns.
// !question

// !cols
