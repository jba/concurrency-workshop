package main

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// heading Passing a value between goroutines

// text
// We can pass a value between goroutines with a WaitGroup.
// !text

func f1() {
	// code
	var wg sync.WaitGroup
	var v int
	wg.Go(func() { v = compute(7) })
	wg.Wait()
	fmt.Println(v)
	// !code
}

func Test_f1(t *testing.T) {
	if stdout(f1) != "49" {
		t.Error("f1 wrong")
	}
}

// text
// But there is a more flexible way: channels.
// !text

// heading Unbuffered channels

// text An unbuffered channel lets two goroutines rendezvous.

func f2() {
	// code
	ch := make(chan int) // create a channel

	go func() { ch <- compute(7) }() // send to ch

	v := <-ch // receive from ch

	fmt.Println(v)
	// !code
}

// text It doesn't matter which happens first, the send or the receive.

func Test_f2(t *testing.T) {
	if stdout(f2) != "49" {
		t.Error("f2 wrong")
	}
}

// heading Multiples

// text
// You can have many senders, and many receivers.
// !text

func f3() {
	// code
	ch := make(chan int)
	for i := range 5 {
		go func() { ch <- compute(i) }()
	}
	for range 5 {
		go func() {
			fmt.Println(<-ch)
		}()
	}
	// TODO: Wait for all goroutines here.
	// !code
}

// // heading Multiples (fixed)

// func f3() {
// 	// code
// 	ch := make(chan int)
// 	for i := range 5 {
// 		go func() { ch <- compute(i) }()
// 	}
// 	var wg sync.WaitGroup
// 	for range 5 {
// 		wg.Go(func() { fmt.Println(<-ch) })
// 	}
// 	wg.Wait()
// 	// !code
// }

// func Test_f3(t *testing.T) {
// 	got := strings.Fields(stdout(f3))
// 	slices.Sort(got)
// 	want := []string{"0", "1", "16", "4", "9"}
// 	if !slices.Equal(got, want) {
// 		t.Errorf("got %v, want %v", got, want)
// 	}
// }

////////////////////////////////////

// heading Timeout, v1

// text
// The `select` statement blocks until one of the cases is ready.
// !text

// problem

func f4() {
	// code
	ch := make(chan int)
	timeout := make(chan bool)
	go func() {
		ch <- compute(7)
	}()
	go func() {
		time.Sleep(20 * time.Millisecond)
		timeout <- true
	}()
	select {
	case v := <-ch:
		fmt.Println(v)
	case <-timeout:
		fmt.Println("timed out")
	}
	// !code
}

// text
// We'll get to the problem after the next slide.
// !text

func Test_f4(t *testing.T) {
	got := stdout(f4)
	if got != "49" {
		t.Errorf("got %q: wrong", got)
	}
}

// heading Timeout, v2

// text Use `time.After` for timeouts.

func f5() {
	// code
	ch := make(chan int)
	go func() {
		ch <- compute(7)
	}()
	select {
	case v := <-ch:
		fmt.Println(v)
		// em
	case <-time.After(20 * time.Millisecond):
		// !em
		fmt.Println("timed out")
	}
	// !code
}

func Test_f5(t *testing.T) {
	got := stdout(f5)
	if got != "49" {
		t.Errorf("got %q: wrong", got)
	}
}

////////////////////////////////////
// heading Goroutine leaks

func f5a() {
	// code
	ch := make(chan int)
	go func() {
		ch <- compute(7)
	}()
	select {
	case v := <-ch:
		fmt.Println(v)
	case <-time.After(20 * time.Millisecond):
		fmt.Println("timed out")
	}
	// !code
}

// question
// - What happens to the first goroutine if there is a timeout?
// - Assume nothing else receives from `ch`
// answer
// 1. `time.After` case executes
// 2. `select` finishes
// 3. goroutine tries to send to `ch`
//
// - The GC does not collect `ch`: there is still a reference to it.
// - The GC does not collect goroutines: they must terminate.
// !question

////////////////////////////////////
// heading Buffered goroutines

// text
// - A channel can have a queue of values.
// - Sending enqueues, blocks if full.
// - Receiving dequeues, blocks if empty.
// - Sender and receiver don't have to rendezvous.
// !text

func f6() {
	// code
	ch := make(chan int, 1) // em , 1
	go func() {
		ch <- compute(7)
	}()
	select {
	case v := <-ch:
		fmt.Println(v)
		// em
	case <-time.After(20 * time.Millisecond):
		// !em
		fmt.Println("timed out")
	}
	// !code
}

// text
// 1. `time.After` case executes
// 2. `select` finishes
// 3. goroutine tries to send to `ch`
// 4. value is enqueued
// 5. goroutine exits
//
// no leaks, no garbage
// !text

////////////////////////////////////
// heading Non-blocking select

// text
// Let's implement this:
// !text

// code
// sendNotification optionally sends a notification without
// blocking the current goroutine.
// To avoid blocking, the notification might be dropped.
func sendNotification(string) {}

// receiveNotification receives a notification if there is one.
// It returns "" if there isn't.
// It never blocks.
func receiveNotification() string { return "" }

// !code

// heading Notifications: first attempt

// code
var nc_1 = make(chan string)

func sendNotification_1(s string) { nc_1 <- s }

func receiveNotification_1() string { return <-nc_1 }

// !code

// text WDYT?

// heading Notifications: solution
// code
var nc_2 = make(chan string, 10) // em , 10

func sendNotification_2(s string) {
	// em
	select {
	case nc_2 <- s:
	default: // if we can't send, drop s
	}
	// !em
}

func receiveNotification_2() string {
	// em
	select {
	case s := <-nc_2:
		return s
	default:
		return ""
	}
}

// !code

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

////////////////////////////////////
// heading Closing channels

// text
// - Close a channel when it will never be sent to again.
// - Every receiver is notified.
// !text

// // code
// type node struct {
// 	val int
// 	left, right *node
// }

// // Return a channel
// func valuesChannel(root *node) chan int
// func sendValues(n *node, ch chan int){
// 	if n == nil {return}
// 	ch <- n.val
// 	sendValues(n.left, ch)
// 	sendValues(n.right, ch)
// }

// 	close(ch)
// 	// !em
// 	}
// 	return ch
// }

// func runCollatz() {
// 	ch := collatzChannel(28)
// }

////////////////////////////////////

func compute(x int) int { return x * x }

func stdout(f func()) string {
	defer func(out *os.File) { os.Stdout = out }(os.Stdout)
	file, err := os.CreateTemp("", "stdout")
	if err != nil {
		panic(err)
	}
	defer os.Remove(file.Name())
	os.Stdout = file
	f()
	if err := file.Close(); err != nil {
		panic(err)
	}
	data, err := os.ReadFile(file.Name())
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(data))
}
