package main

import (
	"fmt"
	"sync"
)

// heading Interleaving

// text
// Goroutines _interleave_ with each other

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

// text
// div.output
// 27357
// !div.output
// !text

// slide
