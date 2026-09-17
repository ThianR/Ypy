package api

import (
	"testing"
	"time"
)

func TestLotExpiryPolicy(t *testing.T) {
	now := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	if err := (Lot{ID: "L1", ItemID: "item", ExpiresAt: now.Add(24 * time.Hour)}).CanSell(now); err != nil {
		t.Fatal(err)
	}
	if err := (Lot{ID: "L2", ItemID: "item", ExpiresAt: now.Add(-time.Hour)}).CanSell(now); err != ErrExpiredLot {
		t.Fatal("expired lot accepted")
	}
}

func TestLotExpiresAtBoundaryAndRequiresIdentity(t *testing.T) {
	now := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	if err := (Lot{ID: "L1", ItemID: "item", ExpiresAt: now}).CanSell(now); err != ErrExpiredLot {
		t.Fatalf("lot sold at expiry boundary: %v", err)
	}
	if err := (Lot{ID: "L2"}).CanSell(now); err != ErrExpiredLot {
		t.Fatalf("lot without item accepted: %v", err)
	}
}
