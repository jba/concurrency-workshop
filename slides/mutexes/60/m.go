package main

// heading Atomics

// div.flex
// code
import (
	"fmt"
	// em
	"sync/atomic"
	// !em
)

type IDGenerator struct {
	prefix string
	// em
	num atomic.Int64
	// !em
}

func NewIDGenerator(prefix string) *IDGenerator {
	return &IDGenerator{prefix: prefix}
}

func (g *IDGenerator) NewID() string {
	// em
	n := g.num.Add(1)
	// !em
	return fmt.Sprintf("%s%d", g.prefix, n)
}

// !code

// text
// Limited operations
//
// Do not combine
// - Two in sequence is not atomic
//
// "These functions require great care to be used correctly."
//
// Recommendation: use only for counters
// !text
// !div.flex
