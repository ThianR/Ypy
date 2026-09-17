package api

import "testing"

func TestMovingAverageCost(t *testing.T) {
	v, err := (Valuation{Quantity: "0", Cost: "0"}).Receive("10", "100")
	if err != nil {
		t.Fatal(err)
	}
	v, err = v.Receive("20", "130")
	if err != nil || v.Quantity != "30.000000" || v.Cost != "120.000000" {
		t.Fatalf("valuation=%#v err=%v", v, err)
	}
}

func TestPendingCostIsExplicit(t *testing.T) {
	if !(Valuation{Quantity: "10", Cost: "100"}).MarkCostPending().PendingCost {
		t.Fatal("pending cost not retained")
	}
}

func TestPendingCostRequiresReasonToResolve(t *testing.T) {
	v := (Valuation{Quantity: "10", Cost: "0"}).MarkCostPending()
	if _, err := v.ResolvePendingCost("100", ""); err != ErrCostAdjustmentRequired {
		t.Fatalf("expected reason requirement, got %v", err)
	}
	resolved, err := v.ResolvePendingCost("100", "factura corregida")
	if err != nil || resolved.Cost != "100.000000" || resolved.PendingCost {
		t.Fatalf("unexpected resolved valuation: %#v %v", resolved, err)
	}
}

func TestMovingAverageRejectsCorruptPreviousState(t *testing.T) {
	if _, err := (Valuation{Quantity: "bad", Cost: "100"}).Receive("1", "10"); err != ErrInvalidCost {
		t.Fatalf("expected corrupt quantity rejection, got %v", err)
	}
	if _, err := (Valuation{Quantity: "2", Cost: "-1"}).Receive("1", "10"); err != ErrInvalidCost {
		t.Fatalf("expected negative cost rejection, got %v", err)
	}
}
