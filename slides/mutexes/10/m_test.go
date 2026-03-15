package m

import (
	"strconv"
	"testing"

	"github.com/jba/concurrency-workshop/internal/testhelp"
)

func TestMutex(t *testing.T) {
	testhelp.WantStdout(t, "40000", run_1)
}

// run with -race to find data race
func TestClever(t *testing.T) {
	testLess := func(f func()) {
		s := testhelp.Stdout(f)
		got, err := strconv.Atoi(s)
		if err != nil {
			t.Fatal(err)
		}
		want := 40_000
		if got >= want {
			t.Fatalf("got %d, want < %d", got, want)
		}
	}

	testLess(run_1)
	testLess(run_2)
}
