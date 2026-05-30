package lograce

import (
	"os"
	"sync"
	"testing"
)

func Test(t *testing.T) {
	l := NewLogger(func(b []byte) {
		_, _ = os.Stdout.Write(b)
	})
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Go(func() {
			for j := range 10 {
				l.Logf("%d, %d\n", i, j)
			}
		})
	}
	wg.Wait()
}
