package m

import "sync"

// title checklocks

// heading Overview

// text
// Checklocks is a tool developed by Google's gvisor team.

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
// TODO
// !output
