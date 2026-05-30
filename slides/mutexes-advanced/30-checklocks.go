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
// file.go:12:23: must hold a.mu exclusively (&({param:b}.mu)) to call changeBalanceLocked, but not held
// !output

// heading The race detector vs. checklocks
//
// text
// <div class="interleave" style="font-size: 70%">
//
// | -race | checklocks |
// | -- | -- |
// | dynamic | static |
// | program runs 4x slower | no effect on runtime speed |
// | finds all races that happen | can miss some races (see logger exercise) |
// | no code changes | needs annotations |
// | run on any code | only code you can annotate |
// | mutexes can be anywhere | only works for globals and struct fields |
// | only data races, not transaction races | only data races, not transaction races |
// </div>

// !text
