package main

import (
	"fmt"
	"sync"
)

// Code must match previous slide exactly.

// heading The race detector

// text
// Data race:
// - Two goroutines access the same memory
// - At least one writes to it
// - _The accesses aren't synchronized_
// !text

// div.flex
// html <div> <!-- one child for flex -->
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
// html </div>
// html <div> <!-- one child for flex -->
// text
// `go run -race .`
// !text
// output
// ==================
// WARNING: DATA RACE
// Read at 0x000000612e58 by goroutine 7:
//   main.count()
//       jba/repos/github.com/jba/concurrency-workshop/slides/mutexes/10/m.go:26 +0x2c
//   sync.(*WaitGroup).Go.func1()
//       jba/sdk/go1.25.5/src/sync/waitgroup.go:239 +0x5d

// Previous write at 0x000000612e58 by goroutine 8:
//   main.count()
//       jba/repos/github.com/jba/concurrency-workshop/slides/mutexes/10/m.go:26 +0x44
//   sync.(*WaitGroup).Go.func1()
//       jba/sdk/go1.25.5/src/sync/waitgroup.go:239 +0x5d
// !output
// html </div>
// !div.flex
