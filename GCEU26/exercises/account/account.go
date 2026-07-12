// Add a mutex to Account so that all methods are
// safe for use by multiple goroutines.

package account

import "errors"

type Account struct {
	balance int
}

func (a *Account) Balance() int {
	return a.balance
}

func (a *Account) Deposit(amount int) error {
	if amount < 0 {
		return errors.New("deposit amount must be non-negative")
	}
	a.balance += amount
	return nil
}

func (a *Account) Withdraw(amount int) error {
	if amount < 0 {
		return errors.New("withdraw amount must be non-negative")
	}
	if a.balance-amount < 0 {
		return errors.New("insufficient balance")
	}
	a.balance -= amount
	return nil
}
