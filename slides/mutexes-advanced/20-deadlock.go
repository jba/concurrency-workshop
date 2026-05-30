package m

import (
	"sync"
	"time"
)

// title Advanced Mutex Topics

////////////////////////////////////
// heading  Our next subject

// html <img height=500 src="slides/mutexes-advanced/deadlock-rebus.png"/>

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
