package collatz

import (
	"io"
)

func collatzConcurrent(w io.Writer, n int) {
	// TODO: your code goes here
}

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
