package intro

import (
	"bytes"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func Test(t *testing.T) {
	t.Run("main_1", func(t *testing.T) {
		main_1()
	})
	t.Run("main_2", func(t *testing.T) {
		main_2()
	})
	t.Run("main_3", func(t *testing.T) {
		main_2()
	})
	t.Run("collatz", func(t *testing.T) {
		var buf bytes.Buffer
		collatzSequential(&buf, 6)
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

	})
}
