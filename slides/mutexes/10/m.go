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

// text Goroutines share the program's memory.

// cols
// code
var c int

func run() {
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

// nextcol

// html <div style="height: 10vw"></div>

// text One possible output:
// text &nbsp;

// output
// 27357
// !output

// text &nbsp;
// text It should be 40000.

// !cols

////////////////////////////////////
// heading Interleavings

// cols

func f1() {
	var c int
	// code
	c++
	// !code
}

// text is actually
func f2() {
	var R0, c int
	// code
	R0 = c
	R0++
	c = R0
	// !code
}

// Make the column wider.
// html <div style="width: 25vw"></div>

// nextcol
/* text

What we want:

<div class="interleave" style="font-size: 70%">

| G1 | G2 |
| -- | -- |
| c++ |  |
|  | c++ |

</div>

What we might get:

<div class="interleave" style="font-size: 70%">

| G1 | G2 |
| -- | -- |
| R0 = c | R0 = c |
| R0++ | R0++ |
| c = R0 | c = R0 |
</div>

*/
// !cols

////////////////////////////////////////////////
// heading Using a mutex

// cols

// code
var mu sync.Mutex // em

var c_1 int

func run_1() {
	var wg sync.WaitGroup
	wg.Go(count)
	wg.Go(count)
	wg.Wait()
	fmt.Println(c_1)
}

func count_1() {
	for range 20_000 {
		mu.Lock() // em
		c_1++
		mu.Unlock() // em
	}
}

// !code

// nextcol
// text &nbsp;
// text
//
// Only one goroutine between `Lock` and `Unlock`
// (a _critical section_).
//
// The code in the critical section happens _atomically_:
// indivisibly.
//
// A mutex limits interleavings.
//
// The zero mutex is unlocked and ready to use.
// !text
// !cols

////////////////////////////////////////////////
// heading Transactions

// text
// A _transaction_ (in this course): an atomic sequence of operations
// that makes sense for the application.
//
// (Atomic, Consistent and Isolated, but not Durable)
//
// Examples:
// - Money transfer
// - Start fulfilling an order, mark it as in progress
// - Add/remove an element and update the size
// !text
