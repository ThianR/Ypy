package events

import (
	"encoding/json"
	"testing"
)

func TestEnvelopeValidatesContractFixture(t *testing.T) {
	var envelope Envelope
	if err := json.Unmarshal([]byte(`{"event_id":"00000000-0000-0000-0000-000000000010","event_type":"pos.operation.accepted","empresa_id":"1","version":1,"origin":"terminal-1","correlation_id":"00000000-0000-0000-0000-000000000011","occurred_at":"2026-09-16T12:00:00Z","payload":{"operation_id":"00000000-0000-0000-0000-000000000001"}}`), &envelope); err != nil {
		t.Fatal(err)
	}
	if err := envelope.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestEnvelopeRejectsInvalidScopeAndPayload(t *testing.T) {
	e := Envelope{EventID: "bad", CorrelationID: "bad", EmpresaID: "0", Version: 0, Payload: json.RawMessage(`[]`)}
	if err := e.Validate(); err != ErrInvalidEnvelope {
		t.Fatalf("expected invalid envelope, got %v", err)
	}
}
