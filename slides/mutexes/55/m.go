package main

// heading Limit critical section size

// div.flex
// code
import (
	"fmt"
	// em
	"sync"
	// !em
)

type IDGenerator struct {
	prefix string
	mu     sync.Mutex
	num    int
}

func NewIDGenerator(prefix string) *IDGenerator {
	return &IDGenerator{prefix: prefix}
}

func (g *IDGenerator) NewID() string {
	g.mu.Lock()
	g.num++
	g.mu.Unlock()
	return fmt.Sprintf("%s%d", g.prefix, g.num)
}

// !code

// html <div>
// text
// Keep locked regions (_critical sections_) small to avoid contention
//
// `defer` is not always needed or useful
// !text
// question
// Find the bug.
// answer
// html <div class="code"><pre>
// func (g *IDGenerator) NewID() string {
// 	g.mu.Lock()
// 	g.num++
// 	<b>n := g.num</b>
// 	g.mu.Unlock()
// 	return fmt.Sprintf("%s%d", g.prefix, <b>n</b>)
// }

// html </div>
// !div.flex
