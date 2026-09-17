package api

import (
	"testing"
	"time"
)

func TestExpiredSessionCannotCreateContext(t *testing.T) {
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	_, err := (Session{ExpiresAt: now}).Context(now)
	if err != ErrSessionExpired {
		t.Fatal("expired session was accepted")
	}
}

func TestIncompleteSessionCannotCreateContext(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := (Session{ID: "s1", UserID: "u1", EmpresaID: "1", SucursalID: "2", ExpiresAt: now.Add(time.Hour)}).Context(now)
	if err != ErrInvalidSession {
		t.Fatalf("expected invalid session, got %v", err)
	}
}
