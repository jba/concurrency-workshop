package timeout

import (
	"errors"
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

func compute(n int) int {
	time.Sleep(time.Duration(n) * time.Millisecond)
	return n
}

func Test(t *testing.T) {
	t.Run("finish", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			const n = 10
			got, err := computeWithTimeout(n)
			time.Sleep(20 * time.Millisecond) // wait for the timeout goroutine to finish
			synctest.Wait()
			if got != n || err != nil {
				t.Errorf("got (%d, %v), want (%d, nil)", got, err, n)
			}
		})
	})
	t.Run("timeout", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			const n = 21
			got, err := computeWithTimeout(n)
			time.Sleep(1 * time.Millisecond) // wait for compute(21) to finish
			synctest.Wait()
			if got != 0 || err != errTimeout {
				t.Errorf("got (%d, %v), want (0, errTimeout)", got, err)
			}
		})
	})
}
