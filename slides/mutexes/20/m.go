package main

import (
	"fmt"
	"sync"
)

////////////////////////////////////
// heading Data races

// text
// A data race happens when:
// - Two goroutines access the same memory
// - At least one writes to it
// - The accesses aren't synchronized
// !text

// text &nbsp;

// cols

// text same goroutine
func f1() {
	var c int
	// code
	c++
	fmt.Println(c)
	// !code
}

// nextcol

// text different memory
func f2() {
	var c1, c2 int
	var wg sync.WaitGroup
	// code
	wg.Go(func() {
		c1++
	})
	wg.Go(func() {
		c2++
	})
	// !code
	wg.Wait()
}

// nextcol

// text no writes
func f3() {
	var c int
	var wg sync.WaitGroup
	// code
	wg.Go(func() {
		fmt.Println(c)
	})
	wg.Go(func() {
		fmt.Println(c)
	})
	// !code
	wg.Wait()
}

// nextcol
// text data race

func f4() {
	var c int
	var wg sync.WaitGroup
	// code bad
	wg.Go(func() { c++ })
	wg.Go(func() { c++ })
	// !code
	wg.Wait()
}

// !cols

////////////////////////////////////
// heading The race detector

// text Looks for data races while the program is running.

// cols

// Code must match 10/m.go exactly.
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

// nextcol

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

// !cols
