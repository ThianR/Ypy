package api

import "testing"

func TestPaymentApplicationsAreIdempotent(t *testing.T) {
	r := NewReceivable("inv-1", "PYG", 100000)
	if err := r.Apply("pay-1", 40000); err != nil {
		t.Fatal(err)
	}
	if err := r.Apply("pay-2", 60000); err != nil {
		t.Fatal(err)
	}
	if r.Pending() != 0 {
		t.Fatalf("pending=%d", r.Pending())
	}
	if err := r.Apply("pay-2", 1); err != ErrApplicationConflict {
		t.Fatalf("duplicate content conflict was not detected: %v", err)
	}
	if err := r.Apply("pay-3", 1); err != ErrOverpayment {
		t.Fatal("overapplication accepted")
	}
}

func TestPaymentApplicationRejectsInvalidReceivable(t *testing.T) {
	r := NewReceivable("inv-1", "PYG", 100)
	if err := r.Apply("", 10); err != ErrInvalidReceivable {
		t.Fatalf("expected invalid receivable error, got %v", err)
	}
	var nilReceivable *Receivable
	if err := nilReceivable.Apply("pay-1", 10); err != ErrInvalidReceivable {
		t.Fatalf("expected nil receivable error, got %v", err)
	}
}

func TestPaymentApplicationRejectsIDConflict(t *testing.T) {
	r := NewReceivable("inv-1", "PYG", 100)
	if err := r.Apply("pay-1", 40); err != nil {
		t.Fatal(err)
	}
	if err := r.Apply("pay-1", 50); err != ErrApplicationConflict || r.Applied != 40 {
		t.Fatalf("payment conflict accepted: %v applied=%d", err, r.Applied)
	}
}
