package account

import (
	"sync"
	"testing"
	"time"
)

func TestAccount(t *testing.T) {
	var a Account

	if got := a.Balance(); got != 0 {
		t.Errorf("initial balance = %d, want 0", got)
	}

	if err := a.Deposit(100); err != nil {
		t.Fatalf("Deposit(100) = %v, want nil", err)
	}
	if got := a.Balance(); got != 100 {
		t.Errorf("balance after deposit = %d, want 100", got)
	}

	if err := a.Withdraw(30); err != nil {
		t.Fatalf("Withdraw(30) = %v, want nil", err)
	}
	if got := a.Balance(); got != 70 {
		t.Errorf("balance after withdraw = %d, want 70", got)
	}

	if err := a.Deposit(-10); err == nil {
		t.Error("Deposit(-10) = nil, want error")
	}

	if err := a.Withdraw(-10); err == nil {
		t.Error("Withdraw(-10) = nil, want error")
	}

	if err := a.Withdraw(100); err == nil {
		t.Error("Withdraw(100) with balance 70 = nil, want error")
	}
}

func TestAccountDepositConcurrent(t *testing.T) {
	var a Account
	var wg sync.WaitGroup

	// 100 goroutines each deposit 1, 100 times
	for range 100 {
		wg.Go(func() {
			for range 100 {
				a.Deposit(1)
				_ = a.Balance()
			}
		})
	}
	wg.Wait()

	// Without proper locking, some deposits would be lost.
	if got, want := a.Balance(), 10_000; got != want {
		t.Errorf("balance = %d, want %d", got, want)
	}
}

func TestAccountWithdrawConcurrent(t *testing.T) {
	var a Account
	var wg sync.WaitGroup

	a.Deposit(10_000)
	// 100 goroutines each deposit 1, 100 times
	for range 100 {
		wg.Go(func() {
			for range 100 {
				a.Withdraw(1)
				_ = a.Balance()
			}
		})
	}
	wg.Wait()

	// Without proper locking, some deposits would be lost.
	if got, want := a.Balance(), 0; got != want {
		t.Errorf("balance = %d, want %d", got, want)
	}
}

// TestWithdrawTOCTOU demonstrates a Time Of Check-Time Of Use race.
// With the buggy implementation that releases the lock between checking
// the balance and decrementing it, multiple goroutines can all pass the
// balance check and then all withdraw, resulting in a negative balance.
func TestWithdrawTOCTOU(t *testing.T) {
	for range 1000 {
		var a Account
		var wg sync.WaitGroup

		a.Deposit(100)

		// 10 goroutines all try to withdraw the full balance.
		// Only one should succeed; the rest should get "insufficient balance".
		for range 1000 {
			wg.Go(func() {
				time.Sleep(time.Nanosecond)
				a.Withdraw(100)
			})
		}
		wg.Wait()

		// With TOCTOU bug, balance can go negative.
		if bal := a.Balance(); bal < 0 {
			t.Errorf("balance = %d, want >= 0 (TOCTOU race detected)", bal)
			return
		}
	}
}
