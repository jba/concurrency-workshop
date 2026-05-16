package patterns

import "fmt"

// title The actor pattern

////////////////////////////////////
// heading Concept

// text
// Actor: something with behavior and state.
//
// Only the actor can access its own state.
//
// Communication is by message passing.

// image gopher-actor.png

////////////////////////////////////

// code
type Accounts struct {
	balances map[string]int
	mailbox  chan accountsMessage
}

type accountsMessage interface {
	isAccountsMessage()
}

type changeBalance struct {
	name     string
	amount   int
	response chan error
}

func (changeBalance) isAccountsMessage() {}

type getBalance struct {
	name     string
	response chan int
}

func (getBalance) isAccountsMessage() {}

func NewAccounts(names []string) *Accounts {
	a := &Accounts{
		balances: map[string]int{},
		mailbox:  make(chan accountsMessage),
	}
	for _, name := range names {
		a.balances[name] = 0
	}
	go func() {
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
	}()
	return a
}

func (a *Accounts) Deposit(name string, amount int) error {
	response := make(chan error)
	a.mailbox <- changeBalance{name: name, amount: amount, response: response}
	return <-response
}

func (a *Accounts) Withdraw(name string, amount int) error {
	return a.Deposit(name, -amount)
}

func (a *Accounts) Balance(name string) int {
	response := make(chan int)
	a.mailbox <- getBalance{name: name, response: response}
	return <-response
}
