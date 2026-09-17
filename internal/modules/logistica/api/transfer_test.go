package api

import "testing"

func TestTransferDoesNotDoubleMoveStock(t *testing.T) {
	s := NewStock()
	s.Set("A", "item", 10)
	t1 := Transfer{ID: "tr-1", ItemID: "item", Origin: "A", Destination: "B", Quantity: 4}
	if err := s.Transfer(t1); err != nil {
		t.Fatal(err)
	}
	if err := s.Transfer(t1); err != ErrDuplicateTransfer {
		t.Fatal("duplicate transfer accepted")
	}
	if s.Balance("A", "item") != 6 || s.Balance("B", "item") != 4 {
		t.Fatal("stock moved incorrectly")
	}
}

func TestTransferRejectsInvalidRouteOrQuantity(t *testing.T) {
	s := NewStock()
	s.Set("A", "item", 10)
	if err := s.Transfer(Transfer{ID: "tr-1", ItemID: "item", Origin: "A", Destination: "A", Quantity: 1}); err != ErrInvalidTransfer {
		t.Fatalf("expected same-location rejection, got %v", err)
	}
	if err := s.Transfer(Transfer{ID: "tr-2", ItemID: "item", Origin: "A", Destination: "B", Quantity: 0}); err != ErrInvalidTransfer {
		t.Fatalf("expected invalid quantity, got %v", err)
	}
}

func TestTransferRejectsIDConflictWithoutChangingStock(t *testing.T) {
	s := NewStock()
	s.Set("A", "item", 10)
	if err := s.Transfer(Transfer{ID: "tr-1", ItemID: "item", Origin: "A", Destination: "B", Quantity: 4}); err != nil {
		t.Fatal(err)
	}
	err := s.Transfer(Transfer{ID: "tr-1", ItemID: "item", Origin: "A", Destination: "C", Quantity: 4})
	if err != ErrTransferConflict || s.Balance("A", "item") != 6 || s.Balance("C", "item") != 0 {
		t.Fatalf("unexpected transfer conflict: %v A=%d C=%d", err, s.Balance("A", "item"), s.Balance("C", "item"))
	}
}
