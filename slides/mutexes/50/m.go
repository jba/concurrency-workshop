package main

// heading Generating unique IDs, unsafe version

import (
	"fmt"
	"sync"
)

// cols
// code bad
// An IDGenerator generates unique identifiers.
type IDGenerator struct {
	prefix string
	num    int
}

// NewIDGenerator creates an IDGenerator whose
// identifiers begin with prefix.
func NewIDGenerator(prefix string) *IDGenerator {
	return &IDGenerator{prefix: prefix}
}

// NewID generates a unique identifier each time
// it is called.
func (g *IDGenerator) NewID() string {
	g.num++
	return fmt.Sprintf("%s%d", g.prefix, g.num)
}

// !code

// nextcol
// question
// What can happen if two goroutines call `g.NewID()` at the same time?
// answer
// You might get the same ID twice.
// !question
// !cols

//////////////////////////////////////////
// heading Generating unique IDs, safe version

// cols
// code
type IDGenerator_1 struct {
	prefix string
	mu     sync.Mutex // em
	num    int
}

// NewIDGenerator is the same.

func (g *IDGenerator_1) NewID_1() string {
	// em
	g.mu.Lock()
	defer g.mu.Unlock()
	// !em
	g.num++
	return fmt.Sprintf("%s%d", g.prefix, g.num)
}

// !code
// nextcol
// text
// `num` must be synchronized to make `NewID` concurrency-safe
//
// `prefix` is written _before_ all calls to `NewID`
//
// The "mutex hat" convention: declare the mutex field
// above the fields it protects

// !text
// !cols

//////////////////////////////////////////
// heading Limit critical section size

// cols
// code
func (g *IDGenerator_1) NewID_2() string {
	g.mu.Lock()
	g.num++
	g.mu.Unlock()
	return fmt.Sprintf("%s%d", g.prefix, g.num)
}

// !code

// nextcol

// text
// Keep locked regions (_critical sections_) small to avoid contention
//
// `defer` is not always needed or useful
// !text
// question
// Find and fix the bug.
// answer
// html <div class="code"><pre>
// func (g *IDGenerator) NewID() string {
// 	g.mu.Lock()
// 	g.num++
// 	<b>n := g.num</b>
// 	g.mu.Unlock()
// 	return fmt.Sprintf("%s%d", g.prefix, <b>n</b>)
// }
// !question

// !cols
