package api

import "testing"

func TestCashMovementIsIdempotentAndRequiresReason(t *testing.T) {
	b := NewMovementBook()
	m := Movement{ID: "m1", SessionID: "s1", Kind: "WITHDRAWAL", Reason: "retiro autorizado", Amount: 100}
	if err := b.Add(m); err != nil {
		t.Fatal(err)
	}
	if err := b.Add(m); err != ErrDuplicateMovement {
		t.Fatal("duplicate movement accepted")
	}
	if b.Count() != 1 {
		t.Fatal("movement count changed")
	}
	if err := b.Add(Movement{ID: "m2", SessionID: "s1", Kind: "INCOME", Amount: 50}); err != ErrMovementReasonRequired {
		t.Fatal("movement without reason accepted")
	}
}

func TestCashMovementRejectsInvalidIdentity(t *testing.T) {
	b := NewMovementBook()
	if err := b.Add(Movement{ID: "", SessionID: "s1", Kind: "INCOME", Reason: "ok", Amount: 1}); err != ErrInvalidMovement {
		t.Fatalf("expected invalid movement, got %v", err)
	}
	if err := b.Add(Movement{ID: "m2", SessionID: "s1", Kind: "UNKNOWN", Reason: "ok", Amount: 1}); err != ErrInvalidMovement {
		t.Fatalf("expected invalid kind, got %v", err)
	}
}

func TestCashMovementRejectsIDConflict(t *testing.T) {
	b := NewMovementBook()
	if err := b.Add(Movement{ID: "m1", SessionID: "s1", Kind: "INCOME", Reason: " ingreso ", Amount: 10}); err != nil {
		t.Fatal(err)
	}
	if err := b.Add(Movement{ID: "m1", SessionID: "s1", Kind: "INCOME", Reason: "otro", Amount: 10}); err != ErrMovementConflict || b.Count() != 1 {
		t.Fatalf("movement conflict accepted: %v count=%d", err, b.Count())
	}
}
