// Imagine that time.After didn't exist.
// How could you write it with a goroutine and a channel?
package timeout

import (
	"errors"
	"fmt"
	"testing"
	"testing/synctest"
	"time"
)

var errTimeout = errors.New("timed out")

func computeWithTimeout(n int) (int, error) {
	c := make(chan int, 1)
	go func() { c <- compute(n) }()
	select {
	case v := <-c:
		return v, nil
	case <-time.After(20 * time.Millisecond): // REPLACE!
		return 0, errTimeout
	}
}

// compute simulates a computation that lasts n milliseconds,
// then returns n.
func compute(n int) int {
	fmt.Println(1)
	time.Sleep(time.Duration(n) * time.Millisecond)
	fmt.Println(2)
	return n
}

func Test(t *testing.T) {
	t.Run("finish", func(t *testing.T) {
		// Verify that a computation running for less than
		// the timeout finishes.
		synctest.Test(t, func(t *testing.T) {
			const n = 10
			got, err := computeWithTimeout(n)
			if got != n || err != nil {
				t.Errorf("got (%d, %v), want (%d, nil)", got, err, n)
			}
		})
	})
	t.Run("timeout", func(t *testing.T) {
		// Verify that a computation that takes more than
		// the timeout times out.
		synctest.Test(t, func(t *testing.T) {
			const n = 21
			got, err := computeWithTimeout(n)
			time.Sleep(1 * time.Millisecond) // wait for compute(21) to finish (we already waited 20 ms)
			if got != 0 || err != errTimeout {
				t.Errorf("got (%d, %v), want (0, errTimeout)", got, err)
			}
		})
	})
}
