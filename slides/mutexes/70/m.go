package main

import (
	"fmt"
	"sync"
)

// heading Protecting a map

type IDGenerator struct {
	mu   sync.Mutex
	nums map[string]int // prefix to next ID
}

func NewIDGenerator(prefix string) *IDGenerator {
	return &IDGenerator{nums: map[string]int{}}
}

func (g *IDGenerator) NewID(prefix string) string {
	g.mu.Lock()
	n := g.nums[prefix]
	n++
	g.nums[prefix] = n
	g.mu.Unlock()
	return fmt.Sprintf("%s%d", prefix, n)
}
