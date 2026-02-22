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
// We can pass a value between goroutines with a `WaitGroup`.
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
	c := make(chan int) // create a channel

	go func() { c <- compute(7) }() // send to c

	v := <-c // receive from c

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
	c := make(chan int)
	for i := range 5 {
		go func() { c <- compute(i) }()
	}
	for range 5 {
		go func() {
			fmt.Println(<-c)
		}()
	}
	// Wait for all goroutines here.
	// !code
}

// // heading Multiples (fixed)

// func f3() {
// 	// code
// 	c := make(chan int)
// 	for i := range 5 {
// 		go func() { c <- compute(i) }()
// 	}
// 	var wg sync.WaitGroup
// 	for range 5 {
// 		wg.Go(func() { fmt.Println(<-c) })
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

func f4() {
	// code bad
	c := make(chan int)
	timeout := make(chan bool)
	go func() { c <- compute(7) }()
	go func() {
		time.Sleep(20 * time.Millisecond)
		timeout <- true
	}()
	select {
	case v := <-c:
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
	c := make(chan int)
	go func() { c <- compute(7) }()
	select {
	case v := <-c:
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

// cols
func f5a() {
	// code bad
	c := make(chan int)
	go func() { c <- compute(7) }()
	select {
	case v := <-c:
		fmt.Println(v)
	case <-time.After(20 * time.Millisecond):
		fmt.Println("timed out")
	}
	// program continues
	// !code
}

// nextcol
// question
// What happens to the first goroutine if there is a timeout?
// answer
// 1. `time.After` case executes
// 2. `select` finishes
// 3. goroutine tries to send to `c`
//
// - The GC does not collect `c`: there is still a reference to it.
// - The GC does not collect goroutines: they must terminate.
// !question

// !cols

////////////////////////////////////
// heading Buffered goroutines

// cols

func f6() {
	// code
	c := make(chan int, 1) // em , 1
	go func() { c <- compute(7) }()
	select {
	case v := <-c:
		fmt.Println(v)
	case <-time.After(20 * time.Millisecond):
		fmt.Println("timed out")
	}
	// program continues
	// !code
}

// text
// - A channel can have a queue of values.
// - Sending enqueues, blocks if full.
// - Receiving dequeues, blocks if empty.
// - Sender and receiver don't have to rendezvous.
// !text

// nextcol

// question
// And now?
// answer
// 1. `time.After` case executes
// 2. `select` finishes
// 3. goroutine tries to send to `c`
// 4. <span style="color:purple">value is enqueued</span>
// 5. goroutine exits
//
// no leaks, no garbage
// !question
// !cols

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

// text What happens here?

// text What if we add buffering?

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

// text Close a channel when it will never be sent to again.

// cols
// code weak
type node struct {
	val         int
	left, right *node
}

func sendValues(n *node, ch chan int) {
	if n == nil {
		return
	}
	sendValues(n.left, ch)
	ch <- n.val
	sendValues(n.right, ch)
}

// !code
// text (In modern Go, we would use an iterator.)
// nextcol
// code weak
func printTree(root *node) {
	c := make(chan int)
	go func() {
		sendValues(root, c)
		// em
		close(c)
		// !em
	}()

	for {
		v, ok := <-c // distinguish closed from zero value // em ok
		if !ok {
			// c is closed
			break
		}
		fmt.Println(v)
	}
}

// !code
// !cols

// heading for...range with a channel

// code
func printTree_1(root *node) {
	c := make(chan int)
	go func() {
		sendValues(root, c)
		close(c)
	}()

	// em
	for v := range c {
		// !em
		fmt.Println(v)
	}
}

// !code

var aTree = &node{
	val:  2,
	left: &node{val: 1},
	right: &node{
		val:   4,
		left:  &node{val: 3},
		right: &node{val: 5},
	},
}

func TestPrintTree(t *testing.T) {
	wantStdout(t, "1\n2\n3\n4\n5", func() { printTree(aTree) })
}

////////////////////////////////////
// heading close broadcasts

// text `close` affects every receiver

// // cols
// // code
// func printTree_2(root *node) {
// 	c := make(chan int)
// 	go func() {
// 		sendValues(root, c)
// 		close(c)
// 	}()

// 	var wg sync.WaitGroup
// 	wg.Go(func() {
// 		for v := range c {
// 			fmt.Println(v)
// 		}
// 	})

// 	for v := range c {
// 		fmt.Println(v)
// 	}
// 	wg.Wait()
// }

// // !code

// // question What will this function do?
// // answer
// // - Print all the values of the tree in some order.
// // - Then return.

// // nextcol

// // code
// func printTree_3(root *node) {
// 	c := make(chan int)
// 	go func() {
// 		sendValues(root, c)
// 		c <- -1 // signal done with -1
// 	}()

// 	var wg sync.WaitGroup
// 	wg.Go(func() {
// 		for v := range c {
// 			fmt.Println(v)
// 		}
// 	})

// 	for v := range c {
// 		fmt.Println(v)
// 	}
// 	wg.Wait()
// }

// // !code

// func TestPrintTree2(t *testing.T) {
// 	got := stdout(func() { printTree_2(aTree) })
// 	fmt.Println(got)
// }

// cols
func send1() {
	// code bad
	c := make(chan int)
	var wg sync.WaitGroup
	wg.Go(func() { c <- 1 }) // em c <- 1
	wg.Go(func() { fmt.Println(<-c) })
	wg.Go(func() { fmt.Println(<-c) })
	wg.Wait()
	// !code
}

// question What does this do?
// answer Print 1, then hang.
// nextcol
func send2() {
	// code
	c := make(chan int)
	var wg sync.WaitGroup
	wg.Go(func() { close(c) }) // em close\(c\)
	wg.Go(func() { fmt.Println(<-c) })
	wg.Go(func() { fmt.Println(<-c) })
	wg.Wait()
	// !code
}

// question What does this do?
// answer Print 0 twice, then finish.

// !cols

func TestSend(t *testing.T) {
	// t.Run("send1", func(t *testing.T){
	// TODO: use synctest
	// })
	t.Run("send2", func(t *testing.T) {
		wantStdout(t, "0\n0", send2)
	})
}

////////////////////////////////////

func compute(x int) int { return x * x }

func wantStdout(t *testing.T, want string, f func()) {
	t.Helper()
	got := stdout(f)
	if got != want {
		t.Errorf("\ngot  %s\nwant %s", got, want)
	}
}

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
