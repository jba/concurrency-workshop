package account

import (
	"errors"
	"sync"
)

type Account struct {
	mu      sync.Mutex
	balance int
}

func (a *Account) Balance() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.balance
}

func (a *Account) Deposit(amount int) error {
	if amount < 0 {
		return errors.New("deposit amount must be non-negative")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.balance += amount
	return nil
}

func (a *Account) Withdraw(amount int) error {
	if amount < 0 {
		return errors.New("withdraw amount must be non-negative")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.balance-amount < 0 {
		return errors.New("insufficient balance")
	}
	a.balance -= amount
	return nil
}

// Not part of the solution; to be discussed during the workshop.

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
