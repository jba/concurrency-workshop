package main

// heading Enter the mutex

// div.flex
// code
import (
	"fmt"
	// em
	"sync"
	// !em
)

// em
var mu sync.Mutex

// !em
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
		// em
		mu.Lock()
		// !em
		c++
		// em
		mu.Unlock()
		// !em
	}
}

// !code
// text
//
// Only one goroutine between `Lock` and `Unlock`
//
// A mutex limits interleavings
// !text
// !div.flex
