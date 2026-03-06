package main

import (
	"fmt"
	"sync"
)

////////////////////////////////
// heading Interleaving

// text
// Goroutines _interleave_ with each other.
// !text

func f1() {
	// code
	var c int

	count := func() {
		for range 20_000 {
			c++
		}
	}

	var wg sync.WaitGroup
	wg.Go(count)
	wg.Go(count)
	wg.Wait()
	fmt.Println(c)
	// !code
}

// output
// 27357
// !output
