// TODO: don't copy mutexes

package m

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

// title Introduction to Synchronization
// subtitle
// Demystifying Concurrency

// GopherCon Europe 2026
// !subtitle

// heading Preparation
// text
// Clone https://github.com/jba/concurrency-workshop
// !text

///////////////////////////////////
// heading Sharing memory

// text Goroutines share the program's memory.

// cols
// code large
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

// output
// 27357
// !output

// text &nbsp;
// text It should be 40000.

// !cols

////////////////////////////////////
// heading Interleavings

// text The scheduler interleaves goroutine executions.
// text - any machine instruction
// text - many possibilities
// text - non-deterministic

////////////////////////////////////
// heading Interleaved increment

// cols

// code
var c_4 int // global variable in memory
// !code

func f1() {
	// code
	c++
	// !code
}

// text is actually
func f2() {
	// code
	var R0 int // register 0 (goroutine-local)
	R0++
	c = R0
	// !code
}

// Make the column wider.
// html <div style="width: 25vw"></div>

// nextcol
/* text

What we want (G1, G2 are goroutines):

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
// heading Race condition

// text
// A _race condition_ is when your program
// allows interleavings that are incorrect.
// !text

////////////////////////////////////////////////
// heading  Which brings us to...

// html <img height=500 src="slides/mutexes/mutexes.png"/>

////////////////////////////////////////////////
// heading Mutexes
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
// html <div style="height: 4vw"></div>
// text A mutex provides <i>mut</i>ual <i>ex</i>clusion.
// text The zero mutex is unlocked and ready to use.
// text
// Only one goroutine allowed between
// `Lock` and `Unlock`
// (a _critical section_).

// The code in the critical section happens _atomically_:
// indivisibly.
// !text
// !cols
//
////////////////////////////////////////////////
// heading How to think about a mutex

// cols

// code
var mu_1 sync.Mutex // em

var c_3 int

func run_2() {
	var wg sync.WaitGroup
	wg.Go(count_1)
	wg.Go(count_1)
	wg.Wait()
	fmt.Println(c_3)
}

func count_3() {
	for range 20_000 {
		mu_1.Lock() // em
		c_3++
		mu_1.Unlock() // em
	}
}

// !code

// nextcol
// html <div style="height: 4vw"></div>
//
// text
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

// //////////////////////////////////////////////
// heading Goroutines and memory
//
// text
// - Global variables shared among all goroutines
// - Struct fields shared among all goroutines accessing
// the struct.
// - Local variables, arguments and named return values
// visible only to the goroutine running that function.
// !text
// code
func F(x int) int {
	x++ // no race condition
	return x
}

// !code

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
// - Reconcile registers (CPU-local storage) with main memory
// - Don't reorder
// !text

////////////////////////////////////////////////
// heading Synchronize all accesses

// text Synchronize all reads and writes to a piece of memory

// code large
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
// | x₁ = c | x₂ = c |
// | x₁++ | x₂++ |
// | c = x₁ | c = x₂ |
//
// </div>
// !question
//
// !cols

////////////////////////////////////////////////
// heading Transactions

// text
// A _transaction_ (in this course): an atomic sequence of operations
// that makes sense for the application.
//
// (Atomic, Consistent and Isolated, but not Durable)
//
// Examples:
// - Money transfer
// - Fulfill an order, mark it as done
// - Add/remove an element and update the size
// - Check a precondition, then take an action
// !text

////////////////////////////////////////////////
// heading Our story so far

// text
// Data races are about low-level memory access.<br/>
// Every data race is a race (almost).
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

// //////////////////////////////////////////////
// heading Another example
// cols
// code
func (a *Account) WithdrawTOCTOU(amount int) error {
	if amount < 0 {
		return errors.New("withdraw amount must be non-negative")
	}
	a.mu.Lock()
	bal := a.balance
	a.mu.Unlock()
	if bal-amount < 0 {
		return errors.New("insufficient balance")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.balance -= amount
	return nil
}

// !code

// nextcol

// question Does this code have a data race?
// answer
// No: every memory access is protected by a mutex.
// !question

// question
// Does it have a race?
// answer
// Yes: the balance can go negative.
// !question

// question
// Is a TOCTOU an exotic bird?
// answer
// Sadly, no. That would have been cool.<br/>
// It stands for Time of Check, Time of Use.
// !question
//
// !cols

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

//////////////////////////////////////////
// heading Atomics

// cols
// text
// Package `sync/atomic` exposes CPU atomic operations
//
// "These functions require great care to be used correctly."
//
// Faster than mutexes, but much more dangerous.

// - Limited operations
// - Sequences of atomics are _not_ atomic
//
// <br/>
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

// ////////////////////////////////////////
// heading IDGenerator with a map

// code bad
// IDGenerator generates unique IDs with different prefixes.
type IDGenerator_m1 struct {
	nums map[string]int // prefix to next ID // em
}

func NewIDGenerator_m1() *IDGenerator_m1 {
	return &IDGenerator_m1{nums: map[string]int{}}
}

func (g *IDGenerator_m1) NewID_m1(prefix string) string { // em prefix string
	n := g.nums[prefix]
	n++
	g.nums[prefix] = n
	return fmt.Sprintf("%s%d", prefix, n)
}

// !code

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

// question
// Concurrency-safe maps wouldn't help here.
// Why not?
// answer
// The read-increment-write section is a transaction.
// !question

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

////////////////////////////////////
// heading Some miscellaneous points
//
// text
// <br/>
// !text

//////////////////////////////////////////
// heading Copy defensively
//
// html <p style="margin-bottom:0">Copy private mutable data</p>

// code
type Server_1 struct {
	mu       sync.Mutex
	sessions []*ServerSession
}

// !code

// text `Server.sessions` is managed by `Server`

// html <p style="margin-bottom:0">It doesn't want anyone else to change it.</p>
//
// code
// Sessions returns a copy of the server's sessions.
func (s *Server) Sessions() []*ServerSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.sessions)
}

// !code

// text From The [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk/blob/4cdbaaf27132e5356ba13973ae50da4edfa876bb/mcp/server.go)

//////////////////////////////////////////
// heading Avoid locking during I/O

// text I/O is slow.
// text Network peers can be _arbitrarily_ slow.
// text Sometimes you have to copy.

// code
func (s *Server) notifySessions(n string) {
	s.mu.Lock() // required to access s.sessions
	sessions := slices.Clone(s.sessions)
	// ...
	s.mu.Unlock()
	// Do I/O with no locks held.
	notifySessions(sessions /* ... */)
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

// code
type Logger struct {
	outMu sync.Mutex
	out   io.Writer
	// ...
}

// Printf prints a lone to the logger.
func (l *Logger) Printf(format string, v ...any) {
	buf := fmt.Appendf(nil, format, v...)
	// ... add newline if needed
	// In general, io.Writer.Write is not guaranteed to write its argument atomically.
	// Put it in a critical section to make sure that log output is not interlevaed.
	l.outMu.Lock()
	defer l.outMu.Unlock()
	_, _ = l.out.Write(buf)
}

// !code

// text Adapted from the [log package](https://github.com/golang/go/blob/e30e65f7a8bda0351d9def5a6bc91471bddafd3d/src/log/log.go)

////////////////////////////////
// heading Beware sharing memory!
//
// text Find the bug.

// link ../../../exercises/logger/logger.go Code
// html <br/><br/><br/>
// link ../../../exercises/logger/solution/logger.go Solution

////////////////////////////////////
// heading  And now ...

// html <img height=500 src="slides/mutexes/deadlock-rebus.png"/>

////////////////////////////////////
// heading Deadlock

// text Goroutines can't make progress because they block each other.

// text Or a single goroutine blocks itself.

////////////////////////////////////
// heading Self-deadlock: example

// cols

type Account_d struct {
	mu      sync.Mutex
	balance int
}

// code
func (a *Account_d) Deposit(amount int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.balance += amount
}

func (a *Account_d) Withdraw(amount int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.balance -= amount
}

func (a *Account_d) TransferTo(b *Account_d, amount int) {
	// Must happen atomically.
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Withdraw(amount)
	b.Deposit(amount)
}

// !code

// nextcol

// text
// <div class="interleave" style="font-size: 70%">
//
// | G |
// | -- |
// | astrid.TransferTo(baxter, 100) |
// | astrid.mu.Lock() |
// | astrid.Withdraw(100) |
// | astrid.mu.Lock() |
// | DEADLOCK |
// </div>
// !text

// !cols

// //////////////////////////////////
// heading Solution: refactor

type Account_d2 struct {
	mu      sync.Mutex
	balance int
}

// cols
// code
func (a *Account_d2) Deposit(amount int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.changeBalanceLocked(amount)
}

func (a *Account_d2) Withdraw(amount int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.changeBalanceLocked(-amount)
}

// The caller must hold a.mu.
func (a *Account_d2) changeBalanceLocked(amount int) {
	a.balance += amount
}

// !code
// nextcol
// code
func (a *Account_d2) TransferTo(
	b *Account_d2, amount int) {
	// Must happen atomically.
	a.mu.Lock()
	defer a.mu.Unlock()
	a.changeBalanceLocked(-amount)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.changeBalanceLocked(amount)
}

// !code
// !cols

////////////////////////////////////
// heading A real-world example

// line Modified from net/http/transport.go

type connLRU struct {
	m map[any]any
}

func (connLRU) remove(any) {}

type pConn struct {
	t            *Transport
	isClientConn bool
	idleTimer    *time.Timer
}
type pConn_1 struct {
	t            *Transport_1
	isClientConn bool
	idleTimer    *time.Timer
}

func (pc *pConn) close(err error)   {}
func (pc *pConn_1) close(err error) {}

var errIdleConnTimeout error

// cols
// code small
type Transport struct {
	idleMu  sync.Mutex
	idleLRU connLRU
	// ...
}

func (t *Transport) removeIdleConn1(pconn *pConn) bool {
	t.idleMu.Lock() // em
	defer t.idleMu.Unlock()
	if pconn.idleTimer != nil {
		pconn.idleTimer.Stop()
	}
	t.idleLRU.remove(pconn) // em t.idleLRU
	// elide
	return false
	// !elide
}

// !code

// nextcol
// code small
func (t *Transport) removeIdleConn(pconn *pConn) bool {
	if pconn.isClientConn {
		return true
	}
	return t.removeIdleConn1(pconn) // em removeIdleConn1
}

// !code

// code small
func (pc *pConn) closeConnIfStillIdle() {
	t := pc.t
	t.idleMu.Lock() // em
	defer t.idleMu.Unlock()
	if _, ok := t.idleLRU.m[pc]; !ok { // em t.idleLRU
		return
	}
	t.removeIdleConn1(pc) // em removeIdleConn1
	pc.close(errIdleConnTimeout)
}

// !code
// !cols

////////////////////////////////////
// heading Solution: refactor

// line Actually from net/http/transport.go

// cols
// code small
type Transport_1 struct {
	idleMu  sync.Mutex
	idleLRU connLRU
	// ...
}

// t.idleMu must be held.
func (t *Transport_1) removeIdleConnLocked(pconn *pConn_1) bool {
	if pconn.idleTimer != nil {
		pconn.idleTimer.Stop()
	}
	t.idleLRU.remove(pconn)
	// elide
	return false
	// !elide
}

// !code
// nextcol
// code small
func (t *Transport_1) removeIdleConn_1(pconn *pConn_1) bool {
	if pconn.isClientConn {
		return true
	}
	// em
	t.idleMu.Lock()
	defer t.idleMu.Unlock()
	return t.removeIdleConnLocked(pconn)
	// !em
}

// !code
// code small
func (pc *pConn_1) closeConnIfStillIdle() {
	t := pc.t
	t.idleMu.Lock()
	defer t.idleMu.Unlock()
	if _, ok := t.idleLRU.m[pc]; !ok {
		return
	}
	t.removeIdleConnLocked(pc) // em
	pc.close(errIdleConnTimeout)
}

// !code
// !cols

// //////////////////////////////////
// heading Multi-goroutine deadlocks
type Account_d3 struct {
	mu      sync.Mutex
	balance int
}

func (*Account_d3) changeBalanceLocked(any) {}
func (*Account_d3) Deposit(any)             {}

// cols
// line Reminder:
// code
func (a *Account_d3) TransferTo(b *Account_d3, amount int) {
	// Must happen atomically.
	a.mu.Lock()
	defer a.mu.Unlock()
	a.changeBalanceLocked(-amount)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.changeBalanceLocked(amount)
}

// !code
var xena, yuri *Account_d3

// html <br/>
// line Concurrently:
func g() {
	// code nonum
	xena.TransferTo(yuri, 100)
	// !code
	// code nonum
	yuri.TransferTo(xena, 100)
	// !code

}

// nextcol

// text
// <div class="interleave" style="font-size: 70%">
//
// | X&rarr;Y | Y&rarr;X |
// | -- | -- |
// | x.TransferTo(y) | y.TransferTo(x) |
// | **x**.mu.Lock() | **y**.mu.Lock() |
// | x.changeBalanceLocked() | y.changeBalanceLocked() |
// | **y**.mu.Lock() | **x**.mu.Lock() |
// | DEADLOCK | DEADLOCK |
// !text

// html <br/>
// text That is only one interleaving!

// !col

// //////////////////////////////////
// heading Solution 1: coarser granularity

// line Locks protect larger critical sections.

// cols
// code small
type Accounts struct {
	mu       sync.Mutex
	balances map[string]int // account name to balance
}

func (a *Accounts) Balance(name string) int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.balances[name]
}

func (a *Accounts) Deposit(name string, amount int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.changeBalanceLocked(name, amount)
}

func (a *Accounts) Withdraw(name string, amount int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.changeBalanceLocked(name, -amount)
}

// !code
// nextcol
// code small

func (a *Accounts) changeBalanceLocked(name string, amount int) {
	a.balances[name] += amount
}

func (a *Accounts) TransferTo(fromName, toName string, amount int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.changeBalanceLocked(fromName, -amount)
	a.changeBalanceLocked(toName, amount)
}

// !code

// text Either Xena grabs the lock first, or Yuri does.

// question What's the problem with this approach?
// answer
// No concurrency: only one goroutine can work with _any_ account at a time.
// !question
// !cols

////////////////////////////////////
// heading Solution 2: lock ordering

// line If all goroutines obtain locks in the same order, deadlock is impossible.

func (*Account_d4) changeBalanceLocked(any) {}

// cols
// code
type Account_d4 struct {
	mu      sync.Mutex
	balance int
	id      int // unique for each account // em
}

// !code
// nextcol
// code

func (a *Account_d4) TransferTo(b *Account_d4, amount int) {
	// Acquire locks in ID order.
	// em
	if a.id < b.id {
		a.mu.Lock()
		b.mu.Lock()
	} else {
		b.mu.Lock()
		a.mu.Lock()
	}
	// !em
	// Unlock order doesn't matter.
	defer a.mu.Unlock()
	defer b.mu.Unlock()

	a.changeBalanceLocked(-amount)
	b.changeBalanceLocked(amount)
}

// !code

// heading Checklocks

// text
// A tool developed by Google's gvisor team.

// Installation:
// ```
// go install gvisor.dev/gvisor/tools/checklocks/cmd/checklocks@go
// ```
//
// Use:
// ```
// checklocks ./...
// ```
// !text

// ///////////////////////
// heading Example

// code
type Account_cl1 struct {
	mu      sync.Mutex
	balance int
}

func (a *Account_cl1) changeBalanceLocked(amount int) {
	a.balance += amount
}

func (a *Account_cl1) TransferTo(b *Account_cl1, amount int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.changeBalanceLocked(-amount)
	b.changeBalanceLocked(amount)
}

// !code

// /////////////////////
// heading Example

// code
type Account_cl2 struct {
	mu sync.Mutex
	// +checklocks:mu
	balance int
}

func (a *Account_cl2) changeBalanceLocked(amount int) {
	a.balance += amount
}

func (a *Account_cl2) TransferTo(b *Account_cl2, amount int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.changeBalanceLocked(-amount)
	b.changeBalanceLocked(amount)
}

// !code

// output
// file.go:7:4: invalid field access, mu (&({param:a}.mu)) must be locked when accessing balance
// !output

// /////////////////////
// heading Example, v2

// code
type Account_cl3 struct {
	mu      sync.Mutex
	balance int
}

// +checklocks:a.mu
func (a *Account_cl3) changeBalanceLocked(amount int) {
	a.balance += amount
}

func (a *Account_cl3) TransferTo(b *Account_cl3, amount int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.changeBalanceLocked(-amount)
	b.changeBalanceLocked(amount)
}

// !code

// output
// file.go:12:23: must hold a.mu exclusively (&({param:b}.mu)) to call changeBalanceLocked, but not held
// !output

// heading The race detector vs. checklocks
//
// text
// <div class="interleave" style="font-size: 100%">
//
// | -race | checklocks |
// | -- | -- |
// | dynamic | static |
// | program runs 4x slower | no effect on runtime speed |
// | finds all races that happen | can miss some races (memory sharing) |
// | no code changes | needs annotations |
// | run on any code | only code you can annotate |
// | mutexes can be anywhere | only works for globals and struct fields |
// | only data races, not transaction races | only data races, not transaction races |
// </div>

// !text
