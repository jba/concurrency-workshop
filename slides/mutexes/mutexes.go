package m

import (
	"fmt"
	"io"
	"iter"
	"reflect"
	"slices"
	"sync"
	"sync/atomic"
)

// title Introduction to Synchronization
// subtitle
// Demystifying Concurrency

// GopherCon Europe 2026
// !subtitle

// heading Sharing memory

// text Goroutines share the program's memory.

// cols
// code
var c int

func run() {
	var wg sync.WaitGroup
	wg.Go(count)
	wg.Go(count)
	wg.Wait()
	fmt.Println(c)
}

func count() {
	for range 20_000 {
		c++
	}
}

// !code

// nextcol

// html <div style="height: 10vw"></div>

// text One possible output:
// text &nbsp;

// output
// 27357
// !output

// text &nbsp;
// text It should be 40000.

// !cols

////////////////////////////////////
// heading Interleavings

// text The scheduler interleaves goroutine executions.
// text - many possibilities
// text - non-deterministic

// html <div style="height: 4vw"></div>

// cols

func f1() {
	var c int
	// code
	c++
	// !code
}

// text is actually
func f2() {
	var R0, c int
	// code
	R0 = c
	R0++
	c = R0
	// !code
}

// Make the column wider.
// html <div style="width: 25vw"></div>

// nextcol
/* text

What we want:

<div class="interleave" style="font-size: 70%">

| G1 | G2 |
| -- | -- |
| c++ |  |
|  | c++ |

</div>
*/
// html <div style="width: 25vw"></div>

// nextcol
/* text
What we might get:

<div class="interleave" style="font-size: 70%">

| G1 | G2 |
| -- | -- |
| R0 = c | R0 = c |
| R0++ | R0++ |
| c = R0 | c = R0 |
</div>
*/
// !cols

////////////////////////////////////////////////
// heading Using a mutex

// cols

// code
var mu sync.Mutex // em

var c_1 int

func run_1() {
	var wg sync.WaitGroup
	wg.Go(count_1)
	wg.Go(count_1)
	wg.Wait()
	fmt.Println(c_1)
}

func count_1() {
	for range 20_000 {
		mu.Lock() // em
		c_1++
		mu.Unlock() // em
	}
}

// !code

// nextcol
// text &nbsp;
// text The zero mutex is unlocked and ready to use.
// text Only one goroutine between `Lock` and `Unlock`
// (a _critical section_).
//
// text
// The code in the critical section happens _atomically_:
// indivisibly.
//
// A mutex limits interleavings.
//
// This is no longer possible:
// <div class="interleave" style="font-size: 70%">

// | G1 | G2 |
// | -- | -- |
// | R0 = c | R0 = c |
// | R0++ | R0++ |
// | c = R0 | c = R0 |
// </div>
// !text
// !cols

////////////////////////////////////////////////
// heading Transactions

// text
// A _transaction_ (in this course): an atomic sequence of operations
// that makes sense for the application.
//
// (DB people: Atomic, Consistent and Isolated, but not Durable)
//
// Examples:
// - Money transfer
// - Fulfill an order, mark it as done
// - Add/remove an element and update the size
// - Check a precondition, then take an action
// !text

////////////////////////////////////////////////
// heading Exercise: Bank Account

// link ../../../exercises/account/account.go Code
// html <br/><br/><br/>
// link ../../../exercises/account/solution/account.go Solution

////////////////////////////////////////////////
// heading Synchronization: more than interleavings

type Account struct {
	mu      sync.Mutex
	balance int
}

// code
func (a *Account) Balance() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.balance
}

// !code

// text Why does this need a mutex?

// text CPU can read/write an `int` atomically.

////////////////////////////////////////////////
// heading Synchronization is a signal

// text
// Modern multicore CPUs cache
//
// Modern compilers optimize
//
// Synchronization points say: "Don't do that."
//
// <br/>
//
// `mu.Lock()` means:
// - Other threads must wait
// - Coordinate CPU caches
// - Reconcile registers with memory
// - Don't reorder
// !text

////////////////////////////////////////////////
// heading Synchronize all accesses

// text Synchronize all reads and writes to a piece of memory

// code

func (a *Account) Balance_1() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.balance
}

func (a *Account) Deposit(amount int) error {
	// ...
	a.mu.Lock()
	defer a.mu.Unlock()
	a.balance += amount
	return nil
}

// !code

////////////////////////////////////
// heading Data races

// text
// A data race happens when:
// - Two goroutines access the same memory
// - At least one writes to it
// - The accesses aren't synchronized
// !text

// text &nbsp;

// cols

// html <p style="margin-bottom:0">same goroutine</p>
func f5() {
	var c int
	// code
	c++
	fmt.Println(c)
	// !code
}

// nextcol

// html <p style="margin-bottom:0">different memory</p>
func f6() {
	var c1, c2 int
	var wg sync.WaitGroup
	// code
	wg.Go(func() {
		c1++
	})
	wg.Go(func() {
		c2++
	})
	// !code
	wg.Wait()
}

// nextcol

// html <p style="margin-bottom:0">no writes</p>
func f3() {
	var c int
	var wg sync.WaitGroup
	// code
	wg.Go(func() {
		fmt.Println(c)
	})
	wg.Go(func() {
		fmt.Println(c)
	})
	// !code
	wg.Wait()
}

// nextcol
// html <p style="margin-bottom:0">data race</p>
func f4() {
	var c int
	var wg sync.WaitGroup
	// code bad
	wg.Go(func() { c++ })
	wg.Go(func() { c++ })
	// !code
	wg.Wait()
}

// !cols

////////////////////////////////////
// heading The race detector

// text Looks for data races while the program is running.

// cols

// code
var c_2 int

func main() {
	var wg sync.WaitGroup
	wg.Go(count)
	wg.Go(count)
	wg.Wait()
	fmt.Println(c_2)
}

func count_2() {
	for range 20_000 {
		c_2++
	}
}

// !code

// output
// 27357
// !output

// nextcol

// text
// `go run -race .`
// !text
// output
// ==================
// WARNING: DATA RACE
// Read at 0x000000612e58 by goroutine 7:
//   main.count()
//       jba/repos/github.com/jba/concurrency-workshop/slides/mutexes/10/m.go:26 +0x2c
//   sync.(*WaitGroup).Go.func1()
//       jba/sdk/go1.25.5/src/sync/waitgroup.go:239 +0x5d

// Previous write at 0x000000612e58 by goroutine 8:
//   main.count()
//       jba/repos/github.com/jba/concurrency-workshop/slides/mutexes/10/m.go:26 +0x44
//   sync.(*WaitGroup).Go.func1()
//       jba/sdk/go1.25.5/src/sync/waitgroup.go:239 +0x5d
// !output

// !cols

////////////////////////////////////
// heading Exercise

// text Run `go test -race` on your bank account solution.

// text Try it with and without locking `Account.Balance`.

////////////////////////////////////////////////
// heading Let's be clever

// cols

var (
	mu_c sync.Mutex
	c_c  int
)

func run_c() {
	var wg sync.WaitGroup
	wg.Go(count_c)
	wg.Go(count_c)
	wg.Wait()
	fmt.Println(c_c)
}

// code
func count_c() {
	for range 20_000 {
		x := c_c + 1 // em
		mu_c.Lock()
		c_c = x // write is protected // em
		mu_c.Unlock()
	}
}

// !code

// nextcol
// question
// What do we think about this optimization?
// answer
// There is still a data race: a read can happen concurrently with a write.

// <div class="interleave" style="font-size: 70%">
//
// | G1 | G2 |
// | -- | -- |
// | x = c + 1 | |
// | c = x | x = c + 1 |
//
// </div>
// !question
//
// !cols

////////////////////////////////////////////////
// heading Let's be even cleverer!

// cols

var (
	mu_cc sync.Mutex
	c_cc  int
)

func run_cc() {
	var wg sync.WaitGroup
	wg.Go(count_cc)
	wg.Go(count_cc)
	wg.Wait()
	fmt.Println(c_cc)
}

// code
func count_cc() {
	for range 20_000 {
		mu_cc.Lock()
		x := c_cc // read is protected // em
		mu_cc.Unlock()
		x++
		mu_cc.Lock()
		c_cc = x // write is protected // em
		mu_cc.Unlock()
	}
}

// !code

// nextcol
// question
// What do we think about this optimization?
// answer
// There is no data race, but this code is still incorrect:

// <div class="interleave" style="font-size: 70%">
//
// | G1 | G2 |
// | -- | -- |
// | x = c | x = c |
// | x++ | x++ |
// | c = x | c = x |
//
// </div>
// !question
//
// !cols

////////////////////////////////////////////////
// heading Our story so far

// text
// Data races are about low-level memory access.<br/>
// Every data race is a concurrency bug (almost).
//
// But races in general are about transactions.
//
// *Not every race is a data race.*
// !text
//
// question Why doesn't Go make maps concurrency-safe?
// answer
// It would be slower, and _it wouldn't help in many cases_.<br/>
// The runtime doesn't know what your transactions are.
// !question

////////////////////////////////////////////////
// heading Generating unique IDs, unsafe version

// cols
// code bad
// An IDGenerator generates unique identifiers.
type IDGenerator struct {
	prefix string
	num    int
}

// NewIDGenerator creates an IDGenerator whose
// identifiers begin with prefix.
func NewIDGenerator(prefix string) *IDGenerator {
	return &IDGenerator{prefix: prefix}
}

// NewID generates a unique identifier each time
// it is called.
func (g *IDGenerator) NewID() string {
	g.num++
	return fmt.Sprintf("%s%d", g.prefix, g.num)
}

// !code

// nextcol
// question
// What can happen if two goroutines call `g.NewID()` at the same time?
// answer
// You might get the same ID twice.
// !question
// !cols

////////////////////////////////////////////////
// heading Generating unique IDs, safe version

// cols
// code
type IDGenerator_1 struct {
	prefix string
	mu     sync.Mutex // em
	num    int
}

// NewIDGenerator is the same.

func (g *IDGenerator_1) NewID_1() string {
	// em
	g.mu.Lock()
	defer g.mu.Unlock()
	// !em
	g.num++
	return fmt.Sprintf("%s%d", g.prefix, g.num)
}

// !code
// nextcol
// text
// `num` must be synchronized to make `NewID` concurrency-safe
//
// `prefix` is written _before_ all calls to `NewID`
//
// The "mutex hat" convention: declare the mutex field
// above the fields it protects

// !text
// !cols

/////////////////////////////////////////
// heading Limit critical section size

// cols
// code
func (g *IDGenerator_1) NewID_2() string {
	g.mu.Lock()
	g.num++
	g.mu.Unlock()
	return fmt.Sprintf("%s%d", g.prefix, g.num)
}

// !code

// nextcol

// text
// Keep critical sections small to avoid contention
//
// `defer` is not always needed or useful
// !text
// question
// Find and fix the bug.
// answer
// code
func (g *IDGenerator_1) NewID_3() string {
	g.mu.Lock()
	g.num++
	n := g.num // em
	g.mu.Unlock()
	return fmt.Sprintf("%s%d", g.prefix, n) // em \bn\b
}

// !code

// !question

// !cols

//////////////////////////////////////////
// heading Avoid locking during I/O

// text I/O is slow.
// text Network peers can be _arbitrarily_ slow.
// text Sometimes you have to copy.

// code
func (s *Server) notifySessions(n string) {
	s.mu.Lock() // required to access s.sessions
	sessions := slices.Clone(s.sessions)
	s.pendingNotifications[n] = nil
	s.mu.Unlock()
	// Do I/O with no locks held.
	notifySessions(sessions, n, changeNotificationParams[n], s.opts.Logger)
}

// !code

// text From The [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk/blob/4cdbaaf27132e5356ba13973ae50da4edfa876bb/mcp/server.go)

type Server struct {
	mu                   sync.Mutex
	sessions             []*ServerSession
	pendingNotifications map[string]*int
	opts                 struct{ Logger int }
}

type ServerSession int

var changeNotificationParams map[string]int

func notifySessions([]*ServerSession, string, int, int) {}

//////////////////////////////////////////
// heading Avoid locking during I/O...unless you need it

func flog() {
	var l struct {
		outMu sync.Mutex
		out   io.Writer
	}
	var buf *[]byte
	// code
	l.outMu.Lock()
	defer l.outMu.Unlock()
	_, err := l.out.Write(*buf) // avoid interleaved log lines
	// !code
	_ = err
}

// text From The [log package](https://github.com/golang/go/blob/e30e65f7a8bda0351d9def5a6bc91471bddafd3d/src/log/log.go)
//////////////////////////////////////////
// heading Another example of copying

// text We don't know how long the caller will hold on to the iterator.

// code
// Sessions returns an iterator that yields a snapshot of the server sessions.
func (s *Server) Sessions() iter.Seq[*ServerSession] {
	s.mu.Lock()
	clients := slices.Clone(s.sessions)
	s.mu.Unlock()
	return slices.Values(clients)
}

// !code

// text From The [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk/blob/4cdbaaf27132e5356ba13973ae50da4edfa876bb/mcp/server.go)

//////////////////////////////////////////
// heading Atomics

// cols
// text
// Package `sync/atomic` exposes CPU atomic operations
//
// "These functions require great care to be used correctly."
//
// - Faster than mutexes, but much more dangerous.
// - Limited operations
// - Sequences of atomics are _not_ atomic
//
// Recommendation: use only for counters
// !text

// nextcol

// code
type IDGenerator_2 struct {
	prefix string
	num    atomic.Int64 // em
}

// NewIDGenerator is the same.

func (g *IDGenerator_2) NewID_3() string {
	n := g.num.Add(1) // em
	return fmt.Sprintf("%s%d", g.prefix, n)
}

// !code
// !cols

//////////////////////////////////////////
// heading Mutexes and slices

// text Each slice element is a separate memory location.
// text No mutex needed here.

func fslice1() []int {
	// code
	var wg sync.WaitGroup
	s := make([]int, 2)
	wg.Go(func() { s[0] = 1 })
	wg.Go(func() { s[1] = 2 })
	wg.Wait()
	// !code
	return s
}

//////////////////////////////////////////
// heading Mutexes and slices, 2

func fslice2() []int {
	// code
	var wg sync.WaitGroup
	var s []int
	wg.Go(func() { s = append(s, 1) }) // em
	wg.Go(func() { s = append(s, 2) }) // em
	wg.Wait()
	// !code
	return s
}

// question Is a mutex needed here?
// answer
// Yes: there is a data race.
// Both goroutines write to the same location, `s`.
// !question
//////////////////////////////////////////
// heading Mutexes and maps

// cols
// code bad
// IDGenerator generates unique IDs with different prefixes.
type IDGenerator_m1 struct {
	nums map[string]int // prefix to next ID // em
}

func NewIDGenerator_m1(prefix string) *IDGenerator_m1 {
	return &IDGenerator_m1{nums: map[string]int{}}
}

func (g *IDGenerator_m1) NewID_m1(prefix string) string { // em prefix string
	n := g.nums[prefix]
	n++
	g.nums[prefix] = n
	return fmt.Sprintf("%s%d", prefix, n)
}

// !code

// nextcol

// text Maps need synchronization.

// question Why aren't maps like slices?
// answer
// The underlying memory locations can change as the map grows and shrinks.
// !question

// !cols

//////////////////////////////////////////
// heading Mutexes and maps, safely

// cols
// code
type IDGenerator_m2 struct {
	mu   sync.Mutex // em
	nums map[string]int
}

// NewIDGenerator is the same.

func (g *IDGenerator_m2) NewID_m2(prefix string) string {
	g.mu.Lock() // em
	n := g.nums[prefix]
	n++
	g.nums[prefix] = n
	g.mu.Unlock() // em
	return fmt.Sprintf("%s%d", prefix, n)
}

// !code

// nextcol

// html <div style="height: 15vw"></div>

// text Concurrency-safe maps wouldn't help here.
// !cols
////////////////////////////////
// heading Optimizations

// text
// Atomics, as we've seen.
//
// `sync.RWMutex` if there are many more reads than writes.
//
// `sync.Map` if there are very few writes (often just one).

// !text

////////////////////////////////
// heading sync.Map example

// cols

// code
type userTypeInfo struct{} // fields omitted

var userTypeCache sync.Map // map[reflect.Type]*userTypeInfo // em sync.Map

func validUserType(rt reflect.Type) (*userTypeInfo, error) {
	if ui, ok := userTypeCache.Load(rt); ok {
		return ui.(*userTypeInfo), nil
	}

	// Construct a new userTypeInfo and atomically
	// add it to the userTypeCache.
	ut := new(userTypeInfo)
	// ...

	ui, _ := userTypeCache.LoadOrStore(rt, ut)
	return ui.(*userTypeInfo), nil
}

// !code

// nextcol

// text &nbsp;
// text From the [encoding/gob](https://github.com/golang/go/blob/master/src/encoding/gob/type.go) package.
// question What about the race?
// answer
// "If we lose the race, we'll waste a little CPU and create a little garbage
// but return the existing value anyway."
// !question
// !cols

////////////////////////////////
// heading Find the bug
// text TODO find example that Mac found in mcp or jsonschema-go

////////////////////////////////
// heading Exercise

// text TODO?
