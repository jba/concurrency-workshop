package m

import (
	"fmt"
	"sync"
)

////////////////////////////////////////////////
// heading Let's be clever

// cols

var (
	mu_1 sync.Mutex
	c_1  int
)

func main_1() {
	var wg sync.WaitGroup
	wg.Go(count_1)
	wg.Go(count_1)
	wg.Wait()
	fmt.Println(c_1)
}

// code
func count_1() {
	for range 20_000 {
		x := c_1 + 1 // em
		mu_1.Lock()
		c_1 = x // write is protected // em
		mu_1.Unlock()
	}
}

// !code

// nextcol
// question
// What do we think about this optimization?
// answer
// There is still a data race: a read can happen concurrently with a write.

// <div class="interleave" style="font-size: 70%">
//
// | G1 | G2 |
// | -- | -- |
// | x = c + 1 | |
// | c = x | x = c + 1 |
//
// </div>
// !question
//
// !cols

////////////////////////////////////////////////
// heading Let's be even cleverer!

// cols

var (
	mu_2 sync.Mutex
	c_2  int
)

func main_2() {
	var wg sync.WaitGroup
	wg.Go(count_2)
	wg.Go(count_2)
	wg.Wait()
	fmt.Println(c_2)
}

// code
func count_2() {
	for range 20_000 {
		mu_2.Lock()
		x := c_2 // read is protected // em
		mu_2.Unlock()
		x++
		mu_2.Lock()
		c_2 = x // write is protected // em
		mu_2.Unlock()
	}
}

// !code

// nextcol
// question
// What do we think about this optimization?
// answer
// There is no data race, but this code is still incorrect:

// <div class="interleave" style="font-size: 70%">
//
// | G1 | G2 |
// | -- | -- |
// | x = c | x = c |
// | x++ | x++ |
// | c = x | c = x |
//
// </div>
// !question
//
// !cols

////////////////////////////////////////////////
// heading Data races aren't the only concurrency bugs

// text
// Data races are about low-level memory access.<br/>
// Every data race is a concurrency bug.
//
// But also think about _transactions_: atomic sequences.<br/>
// The read and write must happen atomically.
//
// Examples of transactions:
// - Money transfer
// - Add/remove an element and update the size
// - Start fulfilling an order, mark it as in progress
// !text

// question Why doesn't Go make maps concurrency-safe?
// answer
// It would be slower, and _it wouldn't help in many cases_.<br/>
// The runtime doesn't know what your transactions are.
// !question

////////////////////
// heading Exercise: Bank Account
