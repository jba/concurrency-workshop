package m

import (
	"fmt"
	"sync"
)

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
		mu.Lock() // em
		c++
		mu.Unlock() // em
	}
}

// !code

// nextcol
// text
//
// Only one goroutine between `Lock` and `Unlock`
//
// The zero mutex is unlocked and ready to use
//
// A mutex limits interleavings
// !text
// !cols
