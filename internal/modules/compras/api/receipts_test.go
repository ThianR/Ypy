package api

import "testing"

func TestPartialReceiptsDoNotExceedOrder(t *testing.T) {
	o := NewOrder("order-1", "item-1", 50)
	for _, r := range []struct {
		id  string
		qty int64
	}{{"r1", 10}, {"r2", 20}, {"r3", 20}} {
		if err := o.Receive(r.id, r.qty); err != nil {
			t.Fatal(err)
		}
	}
	if o.Received != 50 || o.Pending() != 0 {
		t.Fatalf("received=%d pending=%d", o.Received, o.Pending())
	}
	if err := o.Receive("r3", 1); err != ErrReceiptConflict {
		t.Fatalf("receipt content conflict not detected: %v", err)
	}
	if err := o.Receive("r4", 1); err != ErrOverReceipt {
		t.Fatal("over-receipt accepted")
	}
}

func TestReceiptRejectsIDConflictWithoutIncreasingReceived(t *testing.T) {
	o := NewOrder("order-1", "item-1", 10)
	if err := o.Receive("r1", 3); err != nil {
		t.Fatal(err)
	}
	if err := o.Receive("r1", 4); err != ErrReceiptConflict || o.Received != 3 {
		t.Fatalf("receipt conflict accepted: %v received=%d", err, o.Received)
	}
}

func TestReceiptRejectsInvalidOrderIdentity(t *testing.T) {
	o := NewOrder("order-1", "item-1", 5)
	if err := o.Receive("", 1); err != ErrInvalidOrder {
		t.Fatalf("expected invalid order error, got %v", err)
	}
	var nilOrder *Order
	if err := nilOrder.Receive("r1", 1); err != ErrInvalidOrder {
		t.Fatalf("expected nil order error, got %v", err)
	}
}
