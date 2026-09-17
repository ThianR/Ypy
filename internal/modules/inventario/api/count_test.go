package api

import (
	"sync"
	"testing"
)

func TestPhysicalCountAdjustmentIsSingleUse(t *testing.T) {
	c := &PhysicalCount{ID: "count-1", ItemID: "item", Reference: 10, Counted: 8}
	if c.Difference() != -2 {
		t.Fatal("wrong difference")
	}
	if err := c.Approve("supervisor", "merma autorizada"); err != nil {
		t.Fatal(err)
	}
	if err := c.Approve("supervisor", "otra"); err != ErrAdjustmentAlreadyApplied {
		t.Fatal("adjustment applied twice")
	}
}
func TestCountRequiresAuthorization(t *testing.T) {
	c := &PhysicalCount{Reference: 10, Counted: 8}
	if err := c.Approve("", ""); err != ErrAdjustmentUnauthorized {
		t.Fatal("unauthorized adjustment accepted")
	}
}

func TestCountRejectsWhitespaceAuthorization(t *testing.T) {
	c := &PhysicalCount{ID: "count-1", ItemID: "item", Reference: 10, Counted: 8}
	if err := c.Approve(" supervisor ", "   "); err != ErrAdjustmentUnauthorized {
		t.Fatalf("whitespace reason accepted: %v", err)
	}
	if err := c.Approve(" supervisor ", " merma "); err != nil || c.Reason != "merma" {
		t.Fatalf("reason was not normalized: %#v %v", c, err)
	}
}

func TestCountRejectsInvalidScopeAfterAuthorization(t *testing.T) {
	c := &PhysicalCount{Reference: 10, Counted: 8}
	if err := c.Approve("supervisor", "motivo"); err != ErrInvalidCount {
		t.Fatalf("expected invalid count, got %v", err)
	}
}

func TestCountApprovalIsConcurrencySafe(t *testing.T) {
	c := &PhysicalCount{ID: "count-1", ItemID: "item", Reference: 10, Counted: 8}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); results <- c.Approve("supervisor", "motivo") }()
	}
	wg.Wait()
	close(results)
	success, already := 0, 0
	for err := range results {
		if err == nil {
			success++
		}
		if err == ErrAdjustmentAlreadyApplied {
			already++
		}
	}
	if success != 1 || already != 1 {
		t.Fatalf("unexpected concurrent approvals: success=%d already=%d", success, already)
	}
}
