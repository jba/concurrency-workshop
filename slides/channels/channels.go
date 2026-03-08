// TODO: channel direction types
package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

// title Introduction to Channels
// subtitle
// Demystifying Concurrency

// GopherCon Europe 2026
// !subtitle

// heading Prelude: The Collatz Conjecture

// text
// Starting with an integer _n_, repeat the following:
// - If _n_ is even, divide it by 2.
// - If _n_ is odd, multiply it by 3 and add 1.
//
// Examples:
// 4 -> 2 -> 1
// 6 -> 3 -> 10 -> 5 -> 16 -> 8 -> 4 -> 2 -> 1
// 7 -> 22 -> 11 -> 34 -> 17 -> 52 -> 26 -> 13 -> 40 -> 20 -> 10 -> 5 -> 16 -> 8 -> 4 -> 2 -> 1

// Conjecture: All values end up at 1.
// !text

// heading The collatz function
// code
// collatz returns the number of steps to get to 1 from n using the Collatz
// sequence.
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

// heading Passing a value between goroutines

// text
// We can pass a value between goroutines with a `WaitGroup`.
// !text

func f1() {
	// code
	var wg sync.WaitGroup
	var v int
	wg.Go(func() { v = collatz(7) })
	wg.Wait()
	fmt.Println(v)
	// !code
}

// text
// But there is a more flexible way: channels.
// !text

// heading Unbuffered channels

// text An unbuffered channel lets two goroutines rendezvous.
// text They wait for each other.

func f2() {
	// code
	c := make(chan int) // create a channel

	go func() { c <- collatz(7) }() // send to c

	v := <-c // receive from c

	fmt.Println(v)
	// !code
}

// text It doesn't matter which happens first, the send or the receive.
// text Sends and receives are thread-safe.

// heading Multiples

// text
// You can have many senders, and many receivers.
// !text
// TODO: show a diagram with "ping pong"
func f3() {
	// code
	c := make(chan int)
	for i := range 3 {
		go func() { c <- collatz(i) }()
	}
	for range 3 {
		go func() {
			fmt.Println(<-c)
		}()
	}
	// Wait for all goroutines here.
	// !code
}

////////////////////////////////////
// heading The select statement

// text Task: run a goroutine, timing out after a fixed duration.

// text Use `time.After` for timeouts.

func f5() {
	// code bad
	c := make(chan int)
	go func() { c <- collatz(7) }()
	select {
	case v := <-c:
		fmt.Println(v)
	case <-time.After(20 * time.Millisecond):
		fmt.Println("timed out")
	}
	// !code
}

// text The timeout logic is right, but something else is wrong...

////////////////////////////////////
// heading Goroutine leaks

// cols
func f5a() {
	// code bad
	c := make(chan int)
	go func() { c <- collatz(7) }()
	select {
	case v := <-c:
		fmt.Println(v)
	case <-time.After(20 * time.Millisecond):
		fmt.Println("timed out")
	}
	// program continues
	// !code
}

// nextcol
// question
// What happens to the first goroutine if there is a timeout?
// answer
// 1. `time.After` case executes
// 2. `select` finishes
// 3. goroutine tries to send to `c`
//
// - The GC does not collect `c`: there is still a reference to it.
// - The GC does not collect goroutines: they must terminate.
// !question

// !cols

////////////////////////////////////
// heading Buffered channels

// cols
// text A buffered channel has a queue of values.

func f6() {
	// code
	c := make(chan int, 1) // cap(c) == 1 // em 1
	go func() { c <- collatz(7) }()
	select {
	case v := <-c:
		fmt.Println(v)
	case <-time.After(20 * time.Millisecond):
		fmt.Println("timed out")
	}
	// program continues
	// !code
}

// nextcol

// text
// - The size of the queue is the _capacity_ of the channel.
// - Sending enqueues, blocks if full.
// - Receiving dequeues, blocks if empty.
// - Sender and receiver don't have to rendezvous.
// !text

// question
// Now what happens to the first goroutine if there is a timeout?
// answer
// 1. `time.After` case executes
// 2. `select` finishes
// 3. goroutine tries to send to `c`
// 4. <span style="color:purple">value is enqueued in channel</span>
// 5. goroutine exits
//
// No leaks: goroutine terminates

// No garbage: Nothing refers to `c`

// !question
// !cols

////////////////////////////////////
// heading Exercise: replacing time.After

// text See exercises/timeout

////////////////////////////////////
// heading Non-blocking select

// text
// Let's implement an in-process notification service:
// !text

// code
// sendNotification optionally sends a notification without
// blocking the current goroutine.
// To avoid blocking, the notification might be dropped.
func sendNotification(string) {}

// receiveNotification receives a notification if there is one.
// It returns "" if there isn't.
// It never blocks.
func receiveNotification() string { return "" }

// !code

// heading Notifications: first attempt

// code bad
var nc_1 = make(chan string)

func sendNotification_1(s string) { nc_1 <- s }

func receiveNotification_1() string { return <-nc_1 }

// !code

// question
// Does this implementation satisfy our spec?
// answer
// No. Both functions can block.
// !question

// question What if we add buffering?
// answer
// It helps until the buffer fills (send) or is empty (receive).
// Then they block again.
// !question

///////////////////////////////////////
// heading Notifications: solution

// text
// The default case of a `select` executes when no channel operations
// are ready.
// !text

// code
var nc_2 = make(chan string, 10) // em 10

func sendNotification_2(s string) {
	select {
	case nc_2 <- s:
	default: // if we can't send, drop s // em
	}
}

func receiveNotification_2() string {
	select {
	case s := <-nc_2:
		return s
	// em
	default:
		return ""
		// !em
	}
}

// !code

////////////////////////////////////
// heading Closing channels

// text
// close a channel with the `close` builtin
// receiving always returns the zero value
// sending panics
// !text

func cc() {
	// code
	c := make(chan int, 1)
	c <- 1
	close(c)
	fmt.Println(<-c) // prints 0
	c <- 2           // panics
	// !code
}

////////////////////////////////////
// heading Closed channels

// cols

// text Close a channel when it will never be sent to again.

// code weak
type node struct {
	val         int
	left, right *node
}

func (n *node) values() chan int {
	c := make(chan int)
	go func() {
		sendValues(n, c)
		close(c) // em
	}()
	return c
}

// !code

// nextcol

// text &nbsp;

// code weak
func sendValues(n *node, ch chan int) {
	if n == nil {
		return
	}
	sendValues(n.left, ch)
	ch <- n.val
	sendValues(n.right, ch)
}

// !code

// text In modern Go, we would use an iterator.

// !cols

// heading Ranging over a channel
// cols

// code weak
func printTree(root *node) {
	c := root.values()
	for {
		v, ok := <-c // em ok
		if !ok {
			// c is closed
			break
		}
		fmt.Println(v)
	}
}

// !code
// text Two-value receive distinguishes closed from zero value.

// nextcol

// code
func printTree_1(root *node) {
	c := root.values()
	for v := range c { // em
		fmt.Println(v)
	}
}

// !code

// text for...range with a channel ends when closed

// !cols

var aTree = &node{
	val:  2,
	left: &node{val: 1},
	right: &node{
		val:   4,
		left:  &node{val: 3},
		right: &node{val: 5},
	},
}

////////////////////////////////////
// heading close broadcasts

// text `close` affects every receiver

// // cols
// // code
// func printTree_2(root *node) {
// 	c := make(chan int)
// 	go func() {
// 		sendValues(root, c)
// 		close(c)
// 	}()

// 	var wg sync.WaitGroup
// 	wg.Go(func() {
// 		for v := range c {
// 			fmt.Println(v)
// 		}
// 	})

// 	for v := range c {
// 		fmt.Println(v)
// 	}
// 	wg.Wait()
// }

// // !code

// // question What will this function do?
// // answer
// // - Print all the values of the tree in some order.
// // - Then return.

// // nextcol

// // code
// func printTree_3(root *node) {
// 	c := make(chan int)
// 	go func() {
// 		sendValues(root, c)
// 		c <- -1 // signal done with -1
// 	}()

// 	var wg sync.WaitGroup
// 	wg.Go(func() {
// 		for v := range c {
// 			fmt.Println(v)
// 		}
// 	})

// 	for v := range c {
// 		fmt.Println(v)
// 	}
// 	wg.Wait()
// }

// // !code

// func TestPrintTree2(t *testing.T) {
// 	got := stdout(func() { printTree_2(aTree) })
// 	fmt.Println(got)
// }

// cols
func send1() {
	// code bad
	c := make(chan int)
	var wg sync.WaitGroup
	wg.Go(func() { c <- 0 }) // em c <- 0
	wg.Go(func() { fmt.Println(<-c) })
	wg.Go(func() { fmt.Println(<-c) })
	wg.Wait()
	// !code
}

// question What does this do?
// answer Prints 0, then hangs.
// nextcol
func send2() {
	// code
	c := make(chan int)
	var wg sync.WaitGroup
	wg.Go(func() { close(c) }) // em close\(c\)
	wg.Go(func() { fmt.Println(<-c) })
	wg.Go(func() { fmt.Println(<-c) })
	wg.Wait()
	// !code
}

// question What does this do?
// answer Prints 0 twice, then finishes.

// !cols

////////////////////////////////////
// heading Interrupting goroutines

// cols

func f6a() {
	const n = 7
	// code
	c := make(chan int, 1)
	go func() { c <- collatz(n) }()
	select {
	case v := <-c:
		fmt.Println(v)
	case <-time.After(20 * time.Millisecond):
		fmt.Println("timed out")
	}
	// program continues
	// !code
}

// nextcol

// text
// &nbsp;

// `collatz(n)` will keep running until it's done, consuming resources.

// You can't interrupt an arbitrary goroutine.

// It has to check.
// !text
// !cols

////////////////////////////////////
// heading Asking a goroutine to stop

// cols
func f7() {
	n := 7
	// code
	c := make(chan int, 1)
	done := make(chan struct{}) // no value to send // em

	go func() { c <- collatz_1(n, done) }() // em done
	select {
	case v := <-c:
		fmt.Println(v)
	case <-time.After(20 * time.Millisecond):
		fmt.Println("timed out")
		close(done) // em
	}
	// !code
}

// text
// Fun fact: many real-world `close` calls broadcast when something is
// finished, like this one.
// !text

// nextcol
// code

func collatz_1(n int, done chan struct{}) int { // em done chan struct\{\}
	count := 0
	for n > 1 {
		// em
		select {
		case <-done:
			return -1
		default:
		}
		// !em
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
// !cols

////////////////////////////////////
// heading Contexts

// text Contexts carry values down the call chain.
// text They also carry a "doneness" signal.

// code

func collatz_2(ctx context.Context, n int) (int, error) { // em ctx context.Context, error
	count := 0
	for n > 1 {
		select {
		// em
		case <-ctx.Done(): // closed when done
			return 0, ctx.Err() // non-nil if done
			// !em
		default:
		}
		// elide
		if n%2 == 0 {
			n /= 2
		} else {
			n = 3*n + 1
		}
		count++
		// !elide
	}
	return count, nil
}

// !code

////////////////////////////////////
// heading Contexts for timeouts

// code
func collatzWithTimeout(ctx context.Context, n int, tm time.Duration) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, tm)
	defer cancel()
	return collatz_2(ctx, n)
}

// !code

////////////////////////////////////
// heading Exercise: Hedging

// cols

// text
// Your server has two methods of performing a computation, or
// two backends that it can query for a result.
//
// You could try one at random (a kind of load-balancing):
// !text

// code
func getResult(input string) string {
	if rand.Int()%2 == 0 {
		return method1(input)
	} else {
		return method2(input)
	}
}

// !code
// nextcol
// text
// Or you could try them both in parallel, and take the first result
// you get. This is _hedging_.
//
// Start with `getResult` in exercises/hedging, and modify it to implement hedging.
// - Your `getResult` function should call both `method1` and `method2`
// concurrently.
// - It should return the first result it gets.
// - Before returning, it should cancel the other computation.
// !text

// !cols

func method1(string) string { return "" }
func method2(string) string { return "" }

////////////////////////////////////

func compute(x int) int { return x * x }

func computeALittle(x int) (int, bool) {
	return x + 1, x < 10
}
