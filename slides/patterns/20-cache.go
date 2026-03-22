package patterns

import (
	"sync"
)

////////////////////////////////////
// heading A concurrent function cache

// text
// Memoize a function: cache results of previous calls.
// !text

// code bad
// Memo memoizes function calls.
// It is safe for use by multiple goroutines, if the function is.
type Memo[In comparable, Out any] struct {
	f func(In) Out
	m map[In]Out
}

func NewMemo[In comparable, Out any](f func(In) Out) *Memo[In, Out] {
	return &Memo[In, Out]{f: f, m: map[In]Out{}}
}

func (m *Memo[In, Out]) Call(in In) Out {
	out, ok := m.m[in]
	if !ok {
		out = m.f(in)
		m.m[in] = out
	}
	return out
}

// !code

////////////////////////////////////
// heading Safe but not concurrent

// code weak
// Memo memoizes function calls.
// It is safe for use by multiple goroutines, if the function is.
type Memo_2[In comparable, Out any] struct {
	f  func(In) Out
	mu sync.Mutex // em
	m  map[In]Out
}

func NewMemo_2[In comparable, Out any](f func(In) Out) *Memo_2[In, Out] {
	return &Memo_2[In, Out]{f: f, m: map[In]Out{}}
}

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
// heading Safe and concurrent

// code
// Memo memoizes function calls.
// It is safe for use by multiple goroutines, if the function is.
type Memo_3[In comparable, Out any] struct {
	f  func(In) Out
	mu sync.Mutex
	m  map[In]entry[Out] // em
}

func NewMemo_3[In comparable, Out any](f func(In) Out) *Memo_3[In, Out] {
	return &Memo_3[In, Out]{f: f, m: map[In]entry[Out]{}}
}

type entry[Out any] struct {
	out   Out
	waitc chan struct{}
}

func (m *Memo_3[In, Out]) Call(in In) Out {
	m.mu.Lock()
	e, ok := m.m[in]
	if !ok {
		// Haven't seen this input before.
		// Let other goroutines know that this one is working on it.
		e = entry[Out]{waitc: make(chan struct{})}
		m.m[in] = e
	}
	m.mu.Unlock()
	if ok {
		<-e.waitc
		m.mu.Lock()
		defer m.mu.Unlock()
		return m.m[in].out
	}
	// Run the function without the lock.
	e.out = m.f(in)
	// Update the map entry with the result.
	m.mu.Lock()
	m.m[in] = e
	m.mu.Unlock()
	// Wake up waiting goroutines.
	close(e.waitc)
	return e.out
}

// !code
