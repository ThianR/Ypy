package api

import "testing"

func TestPartialReturnRespectsAccumulatedLimit(t *testing.T) {
	s := NewSaleLine("sale-1", "item-1", 5)
	if err := s.Return("ret-1", 2); err != nil {
		t.Fatal(err)
	}
	if err := s.Return("ret-2", 3); err != nil {
		t.Fatal(err)
	}
	if s.Returnable() != 0 {
		t.Fatalf("returnable=%d", s.Returnable())
	}
	if err := s.Return("ret-2", 1); err != ErrReturnConflict {
		t.Fatalf("return content conflict not detected: %v", err)
	}
	if err := s.Return("ret-3", 1); err != ErrReturnExceedsSale {
		t.Fatal("excess return accepted")
	}
}

func TestReturnRejectsEmptyIdentity(t *testing.T) {
	s := NewSaleLine("sale-1", "item-1", 5)
	if err := s.Return("", 1); err != ErrInvalidReturn {
		t.Fatalf("expected invalid return, got %v", err)
	}
	var nilLine *SaleLine
	if err := nilLine.Return("ret-1", 1); err != ErrInvalidReturn {
		t.Fatalf("expected nil line rejection, got %v", err)
	}
}

func TestReturnRejectsIDConflict(t *testing.T) {
	s := NewSaleLine("sale-1", "item-1", 5)
	if err := s.Return("ret-1", 2); err != nil {
		t.Fatal(err)
	}
	if err := s.Return("ret-1", 3); err != ErrReturnConflict || s.Returned != 2 {
		t.Fatalf("return conflict accepted: %v returned=%d", err, s.Returned)
	}
}

func TestNilSaleLineHasNoReturnableQuantity(t *testing.T) {
	var line *SaleLine
	if line.Returnable() != 0 {
		t.Fatal("nil sale line reported returnable stock")
	}
}
