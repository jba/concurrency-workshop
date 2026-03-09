package main

import (
	"fmt"
	"sync"
)

// title Introduction to Synchronization
// subtitle
// Demystifying Concurrency

// GopherCon Europe 2026
// !subtitle

// heading Sharing memory

// text A program's goroutines can all access the program's memory.
// code
var c int

func main() {
	var wg sync.WaitGroup
	wg.Go(count)
	wg.Go(count)
	wg.Wait()
	fmt.Println(c)
}

func count() {
	for range 20_000 {
		c++
	}
}

// !code

// output
// 27357
// !output
