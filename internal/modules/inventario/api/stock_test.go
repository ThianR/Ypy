package api

import "testing"

func TestReceiptIsAppliedOnce(t *testing.T) {
	l := NewLedger()
	r := Receipt{ID: "receipt-1", ItemID: "item-1", Quantity: 10}
	if err := l.Receive(r); err != nil {
		t.Fatal(err)
	}
	if err := l.Receive(r); err != ErrDuplicateReceipt {
		t.Fatal("duplicate receipt accepted")
	}
	if got := l.Balance("item-1"); got != 10 {
		t.Fatalf("expected balance 10, got %d", got)
	}
}

func TestDifferentReceiptsAccumulate(t *testing.T) {
	l := NewLedger()
	_ = l.Receive(Receipt{ID: "a", ItemID: "item-1", Quantity: 10})
	_ = l.Receive(Receipt{ID: "b", ItemID: "item-1", Quantity: 20})
	if got := l.Balance("item-1"); got != 30 {
		t.Fatalf("expected balance 30, got %d", got)
	}
}

func TestReceiptIdentityConflictDoesNotChangeBalance(t *testing.T) {
	l := NewLedger()
	_ = l.Receive(Receipt{ID: "same", ItemID: "item-1", Quantity: 10})
	if err := l.Receive(Receipt{ID: "same", ItemID: "item-2", Quantity: 11}); err != ErrReceiptConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
	if l.Balance("item-1") != 10 || l.Balance("item-2") != 0 {
		t.Fatal("conflicting receipt changed stock")
	}
}

func TestReceiptRejectsInvalidIdentityOrQuantity(t *testing.T) {
	l := NewLedger()
	if err := l.Receive(Receipt{ID: "", ItemID: "item", Quantity: 1}); err != ErrInvalidReceipt {
		t.Fatalf("expected invalid receipt, got %v", err)
	}
	if err := l.Receive(Receipt{ID: "r", ItemID: "item", Quantity: 0}); err != ErrInvalidReceipt {
		t.Fatalf("expected invalid quantity, got %v", err)
	}
}
