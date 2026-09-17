package audit

import (
	"testing"
)

func TestSensitiveAuditRequiresReasonAndRedactsSecrets(t *testing.T) {
	if _, err := New("op-1", "ANULAR_VENTA", "", nil, nil); err != ErrReasonRequired {
		t.Fatalf("expected reason error, got %v", err)
	}
	entry, err := New("op-1", "ANULAR_VENTA", "cliente solicita", map[string]string{"token": "abc", "estado": "CONFIRMADO"}, nil)
	if err != nil || entry.Before["token"] != "[REDACTED]" || entry.Before["estado"] != "CONFIRMADO" {
		t.Fatalf("unexpected audit entry: %#v, %v", entry, err)
	}
}

func TestAuditRequiresTraceableIdentity(t *testing.T) {
	if _, err := New("", "CAMBIAR", "motivo", nil, nil); err != ErrInvalidEntry {
		t.Fatalf("expected invalid audit identity, got %v", err)
	}
	if _, err := New("op-1", "", "motivo", nil, nil); err != ErrInvalidEntry {
		t.Fatalf("expected invalid audit action, got %v", err)
	}
}

func TestRejectedAuditIsExplicitAndRedacted(t *testing.T) {
	entry, err := RejectedEntry("op-2", "OVERRIDE_DENIED", "sin permiso", map[string]string{"secret": "x"})
	if err != nil || !entry.Rejected || len(entry.After) != 0 || entry.Before["secret"] != "[REDACTED]" {
		t.Fatalf("unexpected rejected audit: %#v %v", entry, err)
	}
}
