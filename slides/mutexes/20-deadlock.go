package m

import (
	"sync"
	"time"
)

////////////////////////////////////
// heading  Our next subject

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
// heading Solution: redesign

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
	a.balance -= amount
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
	b.changeBalanceLocked(amount)
}

// !code
// !cols

////////////////////////////////////
// heading A real-world example

// text From net/http/transport.go

type connLRU struct {
	m map[any]any
}

func (connLRU) remove(any) {}

type persistConn struct {
	t            *Transport
	isClientConn bool
	idleTimer    *time.Timer
}

func (pc *persistConn) close(err error) {}

var errIdleConnTimeout error

// cols
// code
type Transport struct {
	idleMu  sync.Mutex
	idleLRU connLRU
	// ...
}

func (t *Transport) removeIdleConn(pconn *persistConn) bool {
	if pconn.isClientConn {
		return true
	}
	return t.removeIdleConn1(pconn)
}

func (t *Transport) removeIdleConn1(pconn *persistConn) bool {
	t.idleMu.Lock()
	defer t.idleMu.Unlock()
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
// code
func (pc *persistConn) closeConnIfStillIdle() {
	t := pc.t
	t.idleMu.Lock()
	defer t.idleMu.Unlock()
	if _, ok := t.idleLRU.m[pc]; !ok {
		return
	}
	t.removeIdleConn1(pc)
	pc.close(errIdleConnTimeout)
}

// !code
