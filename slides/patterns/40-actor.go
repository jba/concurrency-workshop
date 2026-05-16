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
	mailbox  chan accountsMessage
}

type accountsMessage interface{ isAccountsMessage() }

type getBalance struct {
	name     string   // account name
	response chan int // where to send the result
}

func (getBalance) isAccountsMessage() {}

func (a *Accounts) Balance(name string) int {
	response := make(chan int)
	a.mailbox <- getBalance{name: name, response: response}
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

func (changeBalance) isAccountsMessage() {}

func (a *Accounts) Deposit(name string, amount int) error {
	response := make(chan error)
	a.mailbox <- changeBalance{name: name, amount: amount, response: response}
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
		mailbox:  make(chan accountsMessage),
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
	for msg := range a.mailbox {
		switch m := msg.(type) {
		case changeBalance:
			if a.balances[m.name]+m.amount < 0 {
				m.response <- fmt.Errorf("%s: insufficient funds", m.name)
			} else {
				a.balances[m.name] += m.amount
				m.response <- nil
			}
		case getBalance:
			m.response <- a.balances[m.name]
		}
	}
}

// !code

// //////////////////////////////////
// heading Actor pros and cons

// html <div style="line-height:6rem">
// text
// \+ a single goroutine: no deadlocks or race conditions<br/>
// \+ responses can be asynchronous<br/>
// \+ live code controls access

// &ndash; a lot of code to set up<br/>
// &ndash; lots of runtime overhead
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
	a.mailbox <- getBalance{
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
// If messages are exported, callers can wait for responses.
// !text

// code

func (a *Accounts) SendBalance(name string) chan int {
	r := make(chan int)
	a.mailbox <- getBalance{name: name, response: r}
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

// text The actor goroutine determines how to access the state.

// Example: priorities

// code

func (a *Accounts) run_1() {
	for msg := range a.mailbox {
		switch m := msg.(type) {
		case changeBalance:
			if a.balances[m.name]+m.amount < 0 {
				m.response <- fmt.Errorf("%s: insufficient funds", m.name)
			} else {
				a.balances[m.name] += m.amount
				m.response <- nil
			}
		case getBalance:
			m.response <- a.balances[m.name]
		}
	}

}

// !code
