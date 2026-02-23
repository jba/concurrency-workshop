package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// title Introduction to Channels
// subtitle
// Demystifying Concurrency

// GopherCon Europe 2026
// !subtitle

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
// text Sends and receives are thread-safe.

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

// heading The select statement

// text Task: run a goroutine, timing out after a fixed duration.

// heading Timeout, v1

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

// text red border => something's wrong; we'll discuss soon.

func Test_f4(t *testing.T) {
	got := stdout(f4)
	if got != "49" {
		t.Errorf("got %q: wrong", got)
	}
}

// heading Timeout, v1

// text The part we've seen before.

func f4a() {
	// code bad
	// em
	c := make(chan int)
	// !em
	timeout := make(chan bool)
	// em
	go func() { c <- compute(7) }()
	// !em
	go func() {
		time.Sleep(20 * time.Millisecond)
		timeout <- true
	}()
	select {
	// em
	case v := <-c:
		fmt.Println(v)
		// !em
	case <-timeout:
		fmt.Println("timed out")
	}
	// !code
}

// heading Timeout, v1

// text The timeout part.

func f4b() {
	// code bad
	c := make(chan int)
	// em
	timeout := make(chan bool)
	// !em
	go func() { c <- compute(7) }()
	// em
	go func() {
		time.Sleep(20 * time.Millisecond)
		timeout <- true
	}()
	// !em
	select {
	case v := <-c:
		fmt.Println(v)
	// em
	case <-timeout:
		fmt.Println("timed out")
		// !em
	}
	// !code
}

// heading Timeout, v1

// text
// Each case in a `select` statement is a channel operation:
// a send or a receive.
//
// The `select` blocks until one of the cases is ready.
// !text

func f4c() {
	// code bad
	c := make(chan int)
	timeout := make(chan bool)
	go func() { c <- compute(7) }()
	go func() {
		time.Sleep(20 * time.Millisecond)
		timeout <- true
	}()
	// em
	select {
	case v := <-c:
		fmt.Println(v)
	case <-timeout:
		fmt.Println("timed out")
	}
	// !em
	// !code
}

// heading Timeout, v2

// text Use `time.After` for timeouts.

func f5() {
	// code bad
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
// heading Buffered channels

// cols

func f6() {
	// code
	c := make(chan int, 1) // em 1
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
// Now what happens to the first goroutine if there is a timeout?
// answer
// 1. `time.After` case executes
// 2. `select` finishes
// 3. goroutine tries to send to `c`
// 4. <span style="color:purple">value is enqueued in channel</span>
// 5. goroutine exits
//
// No leaks: goroutine terminates

// No garbage: Nothing refers to `c`

// !question
// !cols

////////////////////////////////////
// heading Non-blocking select

// text
// Let's implement an in-process notification service:
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

// question
// Does this implementation satisfy our spec?
// answer
// No. Both functions can block.
// !question

// question What if we add buffering?
// answer
// It helps until the buffer fills (send) or is empty (receive).
// Then they block again.
// !question

///////////////////////////////////////
// heading Notifications: solution

// text
// The default case of a `select` executes when no channel operations
// are ready.
// !text

// code
var nc_2 = make(chan string, 10) // em 10

func sendNotification_2(s string) {
	select {
	case nc_2 <- s:
	// em
	default: // if we can't send, drop s
		// !em
	}
}

func receiveNotification_2() string {
	select {
	case s := <-nc_2:
		return s
	// em
	default:
		return ""
		// !em
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

// cols
// code weak
type node struct {
	val         int
	left, right *node
}

func (n *node) values() chan int {
	c := make(chan int)
	go func() { sendValues(n, c); close(c) }() // em close\(c\)
	return c
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
// nextcol
// code weak
func printTree(root *node) {
	c := root.values()
	for {
		v, ok := <-c // em ok
		if !ok {
			// c is closed
			break
		}
		fmt.Println(v)
	}
}

// !code

// text
// Close a channel when it will never be sent to again.

// In modern Go, we would use an iterator.

// Two-value receive distinguishes closed from zero value.
// !text

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
// heading Interrupting goroutines

// cols

func f6a() {
	// code
	c := make(chan int, 1)
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

// text
// &nbsp;

// `compute(7)` will keep running until it's done, consuming resources.

// You can't interrupt an arbitrary goroutine.

// It has to check.
// !text
// !cols

////////////////////////////////////
// heading Asking a goroutine to stop

// cols
func f7() {
	// code
	c := make(chan int, 1)
	// em
	done := make(chan struct{}) // no value to send
	// !em
	go func() { c <- compute_1(7, done) }() // em done
	select {
	case v := <-c:
		fmt.Println(v)
	case <-time.After(20 * time.Millisecond):
		fmt.Println("timed out")
		// em
		close(done)
		// !em
	}
	// !code
}

// text
// Fun fact: many real-world `close`s broadcast when something is
// finished, like this one.
// !text

// nextcol
// code
func compute_1(x int, done chan struct{}) int { // em done chan struct\{\}
	t := 0
	for {
		select {
		// em
		case <-done:
			return -1
			// !em
		default:
			x, ok := computeALittle(t)
			if !ok {
				return t
			}
			t += x
		}
	}
}

// !code
// !cols

func TestF7(t *testing.T) {
	wantStdout(t, "15", f7)
}

////////////////////////////////////
// heading Contexts

// cols
func f8(ctx context.Context) {
	// code
	c := make(chan int) // unbuffered
	// em
	ctx, cancel := context.WithTimeout(
		context.Background(),
		20*time.Millisecond)
	defer cancel()
	// !em
	go func() { c <- compute_2(ctx, 7) }() // em ctx
	// em
	fmt.Println(<-c) // prints -1 on timeout
	// !em
	// !code
}

// text
// Use `context.Context` for timeouts.

// Contexts inherit timeouts and cancellations from parents.

// `cancel` must always be called to clean up resources.
// !text

// nextcol
// code
func compute_2(ctx context.Context, x int) int {
	t := 0
	for {
		select {
		// em
		case <-ctx.Done():
			// !em
			return -1
		default:
			x, ok := computeALittle(t)
			if !ok {
				return t
			}
			t += x
		}
	}
}

// !code

// text
// `context.Done` channel closed when context times out or is canceled.
// !text

// !cols

////////////////////////////////////
// heading Contexts for real

// text What a "real" function might look like.

// cols
// code
func computeWithTimeout(ctx context.Context,
	tm time.Duration, arg int,
) (int, error) {
	c := make(chan int, 1)
	ctx, cancel := context.WithTimeout(ctx, tm)
	defer cancel()
	go func() { c <- compute_2(ctx, arg) }()
	select {
	case v := <-c:
		return v, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

// !code

// nextcol

// text &nbsp;

// question Do we still need a buffered channel?
// answer
// Yes. If the second select case is taken, the goroutine would be
// blocked forever sending to `c`, even if `compute` returns early.
// !question

// question Should the `select` have a default case?
// answer
// No. The only two possibilities are that `compute` finishes on time
// and sends to `c`, or that the context times out and closes its `Done`
// channel.
// !question

// !cols

////////////////////////////////////
// heading Contexts and cancellation

// text Use `Context` for cancelling for other reasons too.

// cols
// code
func computeWithCancel(ctx context.Context, arg int) (
	int, error,
) {
	c := make(chan int, 1)
	ctx, cancel := context.WithCancel(ctx) // em context.WithCancel\(.*\)
	defer cancel()
	go func() { c <- compute_2(ctx, arg) }()
	select {
	case v := <-c:
		return v, nil
		// em
	case <-userCancels():
		cancel()
		return 0, ctx.Err()
		// !em
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

// !code
// nextcol

// text
// &nbsp;

// `userCancels` returns a channel that is closed when a button is clicked.

// Still defer `cancel` so it's always called.

// Still select `ctx.Done`: arg context may become done.
// !text

// !cols

func userCancels() chan int { return nil }

////////////////////////////////////
// heading Exercise: Hedging

// text
// Your server has two methods of performing a computation, or
// two backends that it can query for a result.
//
// You could try one at random (a kind of load-balancing):
// !text

// code
func getResult(input string) string {
	if rand.Int()%2 == 0 {
		return method1(input)
	} else {
		return method2(input)
	}
}

// !code

// text
// Or you could try them both in parallel, and take the first result
// you get. This is _hedging_.
//
// Start with the code above, in TODO, and modify it to implement hedging.
// - Your `getResult` function should call both `method1` and `method2`
// concurrently.
// - It should return the first result it gets.
// - Before returning, it should cancel the other computation.
// !text

func method1(string) string { return "" }
func method2(string) string { return "" }

////////////////////////////////////

func compute(x int) int { return x * x }

func computeALittle(x int) (int, bool) {
	return x + 1, x < 10
}

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
