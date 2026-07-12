package collatz

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func Test(t *testing.T) {
	var buf bytes.Buffer
	collatzConcurrent(&buf, 6)
	var got []int
	for s := range strings.FieldsSeq(buf.String()) {
		g, err := strconv.Atoi(s)
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, g)
	}
	want := []int{0, 1, 7, 2, 5, 8}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

}

// This benchmark shows that con128 breaks even
// at 1024 on an 8-CPU (virtual) machine.
//
// > go test -count 6 -bench . > /tmp/bench.out
// > benchstat -row /n -col /p /tmp/bench.out

func Benchmark(b *testing.B) {
	for _, n := range []int{8, 16, 64, 1024} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.Run("p=seq", func(b *testing.B) {
				for b.Loop() {
					collatzSequential(io.Discard, n)
				}
			})
			b.Run("p=con", func(b *testing.B) {
				for b.Loop() {
					collatzConcurrent(io.Discard, n)
				}
			})
			b.Run("p=con128", func(b *testing.B) {
				for b.Loop() {
					collatzConcurrentMore(io.Discard, n)
				}
			})
		})
	}
}

func collatzSequential(w io.Writer, n int) {
	for i := 1; i <= n; i++ {
		fmt.Fprintln(w, collatz(i))
	}
}
