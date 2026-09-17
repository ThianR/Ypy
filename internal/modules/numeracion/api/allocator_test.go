package api

import (
	"testing"
	"time"
)

func TestBlocksAreExclusive(t *testing.T) {
	a := NewAllocator(100)
	one := a.Reserve("terminal-a", 2)
	two := a.Reserve("terminal-b", 2)
	n1, _ := one.Consume("terminal-a")
	n2, _ := two.Consume("terminal-b")
	if n1 != 100 || n2 != 102 {
		t.Fatalf("unexpected numbers: %d %d", n1, n2)
	}
	if _, err := one.Consume("terminal-b"); err != ErrWrongTerminal {
		t.Fatal("foreign terminal consumed block")
	}
}

func TestExhaustedBlockDoesNotReuseNumber(t *testing.T) {
	b := NewAllocator(1).Reserve("terminal-a", 1)
	_, _ = b.Consume("terminal-a")
	if _, err := b.Consume("terminal-a"); err != ErrRangeExhausted {
		t.Fatal("exhausted block reused number")
	}
}

func TestReserveRejectsInvalidRange(t *testing.T) {
	a := NewAllocator(1)
	if a.Reserve("", 1) != nil || a.Reserve("terminal-a", 0) != nil {
		t.Fatal("invalid range was reserved")
	}
}

func TestBlockCannotBeConsumedAfterPermissionExpiry(t *testing.T) {
	at := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	b := NewAllocator(1).ReserveUntil("terminal-a", 1, at)
	if _, err := b.ConsumeAt("terminal-a", at); err != ErrPermissionExpired {
		t.Fatalf("expected expired permission, got %v", err)
	}
}
