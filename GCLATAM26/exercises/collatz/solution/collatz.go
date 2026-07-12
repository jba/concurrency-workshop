package collatz

import (
	"fmt"
	"io"
	"sync"
)

func collatzConcurrent(w io.Writer, n int) {
	results := make([]int, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			results[i] = collatz(i + 1)
		})
	}
	wg.Wait()
	for _, r := range results {
		fmt.Fprintln(w, r)
	}
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

func collatzConcurrentMore(w io.Writer, n int) {
	// Doing more work per goroutine for better performance.
	const chunk = 128
	results := make([]int, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i += chunk {
		wg.Go(func() {
			for j := range chunk {
				if i+j < n {
					results[i+j] = collatz(i + j + 1)
				}
			}
		})
	}
	wg.Wait()
	for _, r := range results {
		fmt.Fprintln(w, r)
	}
}
