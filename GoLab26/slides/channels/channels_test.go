package main

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/jba/concurrency-workshop/internal/testhelp"
)

func TestCollatz(t *testing.T) {
	for _, tc := range []struct {
		n, want int
	}{
		{4, 2},
		{6, 8},
		{7, 16},
	} {
		got := collatz(tc.n)
		if got != tc.want {
			t.Errorf("collatz(%d) = %d, want %d", tc.n, got, tc.want)
		}
	}
}

func TestCollatzWithTimeout(t *testing.T) {
	ctx := context.Background()

	const arg = 75_128_138_247
	const res = 1228

	t.Run("completes", func(t *testing.T) {
		got, err := collatzWithTimeout(ctx, arg, time.Second)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != res {
			t.Errorf("got %d, want %d", got, res)
		}
	})

	t.Run("times out", func(t *testing.T) {
		_, err := collatzWithTimeout(ctx, arg, time.Nanosecond)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal("expected timeout error")
		}
	})
}

func Test_f1(t *testing.T) {
	if testhelp.Stdout(f1) != "16" {
		t.Error("f1 wrong")
	}
}

func Test_f2(t *testing.T) {
	if testhelp.Stdout(f2) != "16" {
		t.Error("f2 wrong")
	}
}

func Test_f5(t *testing.T) {
	got := testhelp.Stdout(f5)
	if got != "16" {
		t.Errorf("got %q: wrong", got)
	}
}

func TestNotifications(t *testing.T) {
	for i := range cap(nc_2) + 1 {
		sendNotification_2(strconv.Itoa(i))
	}
	var want []string
	for i := range cap(nc_2) {
		want = append(want, strconv.Itoa(i))
	}
	want = append(want, "")
	var got []string
	for range cap(nc_2) + 1 {
		got = append(got, receiveNotification_2())
	}
	if !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCC(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic")
		}
	}()
	cc()
}

// func TestCC2(t *testing.T) {
// 	wantStdout(t, "1\n0", cc2)
// }

func TestPrintTree(t *testing.T) {
	testhelp.WantStdout(t, "1\n2\n3\n4\n5", func() { printTree(aTree) })
}

func TestSend(t *testing.T) {
	// t.Run("send1", func(t *testing.T){
	// TODO: use synctest
	// })
	t.Run("send2", func(t *testing.T) {
		testhelp.WantStdout(t, "0\n0", send2)
	})
}

func TestF7(t *testing.T) {
	testhelp.WantStdout(t, "16", f7)
}
