package intro

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// title Understanding Concurrency
// subtitle
// Jonathan Amsterdam
// GopherCon LATAM 2026
// !subtitle

///////////////////////////////////
// heading Who's Jonathan Amsterdam?

// text
// Go team member since 2018
// - pkg.go.dev
// - `log/slog`
// - `http.ServeMux` patterns
// - Go Model Context Protocol SDK
//
// First time in...
// - this conference
// - Florianópolis
// - Brazil
// - South America!
// !text

///////////////////////////////////
// heading Preparation
// text
// Clone https://github.com/jba/concurrency-workshop
//
// We will be working in `GCLATAM26`.
// !text

///////////////////////////////////
// heading The LLElephant in the Room

// img ../../static/llelephant.png

///////////////////////////////////
// heading Concurrency vs. Parallelism

// text
// Parallelism: doing several things at the same time.

// Concurrency: _seeming_ to do several things at the
// same time.
// !text

// heading Concurrency without Parallelism

// text
// <iframe width="560" height="315" src="https://www.youtube.com/embed/Pa0wXSh4I34?si=k2GjH3OSzerAA4hS" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen>
// </iframe>
// !text

///////////////////////////////////
// heading Go and Concurrency

// text
// A _goroutine_ is a unit of concurrency.
//
// - A separate thread of execution
// - A separate program counter + stack
// - A separate "finger" walking through code
//   and executing it
//
// goroutines >> CPUs
//
// The _scheduler_ assigns goroutines to CPUs.
// !text

///////////////////////////////////
// heading Go and Concurrency

// text
// Use the `go` keyword to start a goroutine.
//
// `go` is followed by a function call.
//
// The goroutine runs until the function exits.
//
// All goroutines run concurrently until the main
// goroutine exits.
// !text

func main_1() {
	go fmt.Println(1)
	go fmt.Println(2)
}

// question How many goroutines does this program have?
// answer
// Three: the two that are started with `go`, and the
// main goroutine.
// !question

// question What does this program print?
// answer
// Nothing: the main goroutine exits before the others have
// a chance to run.
// !question

// heading Fix 1

func main_2() {
	go fmt.Println(1)
	go fmt.Println(2)
	time.Sleep(50 * time.Millisecond)
}

// output
// 1
// 2
// !output

// output
// 2
// 1
// !output

// output
// !output

// text goroutine execution is _non-deterministic_.

// heading WaitGroup

// text
// Use a `sync.WaitGroup` to wait for
// goroutines to finish.
// !text

func main_3() {
	var wg sync.WaitGroup
	wg.Go(func() {
		fmt.Println(1)
	})
	wg.Go(func() {
		fmt.Println(2)
	})
	wg.Wait()
}

// output
// 1
// 2
// !output

// output
// 2
// 1
// !output

// text How can we get results in order?

////////////////////////////////////
// heading The Collatz Conjecture

// text
// Starting with an integer _n_, repeat the following:
// - If _n_ is even, divide it by 2.
// - If _n_ is odd, multiply it by 3 and add 1.
//
// Examples:
// - 4 -> 2 -> 1
// - 6 -> 3 -> 10 -> 5 -> 16 -> 8 -> 4 -> 2 -> 1
// - 7 -> 22 -> 11 -> 34 -> 17 -> 52 -> 26 -> 13 -> 40 -> 20 -> 10 -> 5 -> 16 -> 8 -> 4 -> 2 -> 1

// Conjecture: All values end up at 1.
// !text

// heading The collatz function
// code
// collatz returns the number of steps to get to 1 from n
// using the Collatz sequence.
func collatz(n int) int {
	count := 0
	for n > 1 {
		if n%2 == 0 {
			n /= 2
		} else {
			n = 3*n + 1
		}
		count++
	}
	return count
}

// !code

////////////////////////////////////
// heading Computing Collatz(n)

// collatzSequential prints the first n values
// of the collatz function to w in order.
func collatzSequential(w io.Writer, n int) {
	for i := 1; i <= n; i++ {
		fmt.Fprintln(w, collatz(i))
	}
}

////////////////////////////////////
// heading Exercise: Collatz concurrently

// text
// Start with GCLATAM26/exercises/collatz/collatz.go.

// Write collatzConcurrent, which computes
// the values of collatz concurrently, _but
// prints the results in order_.

// Hint: use a slice.

// (You won't need synchronization.)

// !text

////////////////////////////////////
// heading Questions about the exercise

// question Does `collatzConcurrent(10)` run concurrently, or in parallel?
// answer
// Both, on modern hardware. All modern (non-embedded) computers
// have more than one CPU. The Go scheduler will assign goroutines
// to CPUs until there are no more available CPUs.
// (This answer assumes that `collatz(n)` runs slowly enough for small `n`)
// !question

// question Which is faster, the sequential version or the concurrent one?
// answer
// The sequential, by a lot.
// !question

////////////////////////////////////
// heading Concurrency and performance

// text
// The concurrent version will run _slower_ than
// the sequential one.

// Concurrency has overhead.

// To make it worthwhile, you must do a lot of work
// in each goroutine.

// See the benchmark in exercises/collatz/collatz_test.go.
// 128 collatz calls per goroutine does much better.
// !text
