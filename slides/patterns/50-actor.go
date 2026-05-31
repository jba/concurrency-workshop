package patterns

import (
	"fmt"
	"sync"
)

// title The actor pattern

////////////////////////////////////
// heading Concept

// cols
// text
// Actor: something with behavior and state.
//
// Only the actor can access its own state.
//
// Communication is by message passing.
// !text

// nextcol

// image gopher-actor.png

// !cols

// //////////////////////////////////
// heading Balances actor, 1
// code
type Accounts struct {
	balances map[string]int // account name to balance
	getc     chan getBalance
	changec  chan changeBalance
}

type getBalance struct {
	name     string   // account name
	response chan int // where to send the result
}

func (a *Accounts) Balance(name string) int {
	response := make(chan int)
	a.getc <- getBalance{name: name, response: response}
	return <-response
}

// !code

// //////////////////////////////////
// heading Balances actor, 2

// code
type changeBalance struct {
	name     string
	amount   int
	response chan error
}

func (a *Accounts) Deposit(name string, amount int) error {
	response := make(chan error)
	a.changec <- changeBalance{name: name, amount: amount, response: response}
	return <-response
}

func (a *Accounts) Withdraw(name string, amount int) error {
	return a.Deposit(name, -amount)
}

// !code

// //////////////////////////////////
// heading Balances actor, 3

// code
func NewAccounts(names []string) *Accounts {
	a := &Accounts{
		balances: map[string]int{},
		getc:     make(chan getBalance),
		changec:  make(chan changeBalance),
	}
	for _, name := range names {
		a.balances[name] = 0
	}
	go a.run()
	return a
}

// !code

// //////////////////////////////////
// heading Balances actor, 4

// code
func (a *Accounts) run() {
	for {
		select {
		case m := <-a.getc:
			m.response <- a.balances[m.name]
		case m := <-a.changec:
			if a.balances[m.name]+m.amount < 0 {
				m.response <- fmt.Errorf("%s: insufficient funds", m.name)
			} else {
				a.balances[m.name] += m.amount
				m.response <- nil
			}
		}
	}
}

// !code

// //////////////////////////////////
// heading Actor cons and pros

// html <div style="line-height:6rem">
// text
// &ndash; a lot of code to set up<br/>
// &ndash; lots of runtime overhead

// \+ a single goroutine: no deadlocks or race conditions<br/>
// \+ responses can be asynchronous<br/>
// \+ live code controls access
// !text
// html </div>

// //////////////////////////////////
// heading Actor vs. Mutex

// text A single mutex also eliminates deadlocks and races.
// cols
// code
type Accounts_1 struct {
	mu       sync.Mutex
	balances map[string]int
}

func (a *Accounts_1) Balance(name string) int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.balances[name]
}

// !code

// nextcol

// code
func (a *Accounts) Balance_1(name string) int {
	response := make(chan int)
	a.getc <- getBalance{
		name:     name,
		response: response,
	}
	return <-response
}

// !code

// !cols

// //////////////////////////////////
// heading Actor responses can be asynchronous

// text
// By separating sending from receiving, callers can wait for responses.
// !text

// code

func (a *Accounts) SendBalance(name string) chan int {
	r := make(chan int)
	a.getc <- getBalance{name: name, response: r}
	return r // em
}

// !code

func f() {
	doOtherStuff := func() {}
	// code
	accounts := NewAccounts([]string{"Alice", "Bob"})
	ac := accounts.SendBalance("Alice")
	bc := accounts.SendBalance("Bob")
	doOtherStuff()
	abal := <-ac
	bbal := <-bc
	// !code
	_ = abal
	_ = bbal
}

// //////////////////////////////////
// heading Live code controls access

// cols
// text The actor goroutine determines how to access the state.
//
// text For example, we can implement priorities.
//
// nextcol
//
// code small
func (a *Accounts) run_1() { // get has priority
	for {
		// em
		// If a get and change are both waiting, take the get first.
		select {
		case m := <-a.getc:
			m.response <- a.balances[m.name]
			continue
		default:
		}
		// !em
		// Take whichever is ready.
		select {
		case m := <-a.getc:
			m.response <- a.balances[m.name]
		case m := <-a.changec:
			if a.balances[m.name]+m.amount < 0 {
				m.response <- fmt.Errorf("%s: insufficient funds", m.name)
			} else {
				a.balances[m.name] += m.amount
				m.response <- nil
			}
		}
	}
}

// !code
// !cols
