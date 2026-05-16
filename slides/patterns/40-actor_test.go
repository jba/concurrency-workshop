package patterns

import "testing"

func TestAccounts(t *testing.T) {
	a := NewAccounts([]string{"alice", "bob"})

	// Initial balances should be zero
	if got := a.Balance("alice"); got != 0 {
		t.Errorf("initial balance: got %d, want 0", got)
	}

	// Deposit
	if err := a.Deposit("alice", 100); err != nil {
		t.Errorf("Deposit: %v", err)
	}
	if got := a.Balance("alice"); got != 100 {
		t.Errorf("after deposit: got %d, want 100", got)
	}

	// Withdraw
	if err := a.Withdraw("alice", 30); err != nil {
		t.Errorf("Withdraw: %v", err)
	}
	if got := a.Balance("alice"); got != 70 {
		t.Errorf("after withdraw: got %d, want 70", got)
	}

	// Withdraw insufficient funds
	if err := a.Withdraw("alice", 100); err == nil {
		t.Error("Withdraw with insufficient funds: want error, got nil")
	}
	// Balance should be unchanged
	if got := a.Balance("alice"); got != 70 {
		t.Errorf("after failed withdraw: got %d, want 70", got)
	}

	// Other account is independent
	if got := a.Balance("bob"); got != 0 {
		t.Errorf("bob's balance: got %d, want 0", got)
	}
}
