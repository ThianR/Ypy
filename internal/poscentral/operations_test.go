package poscentral

import "testing"

func TestApplyIsIdempotent(t *testing.T) {
	s := NewStore()
	op := Operation{ID: "00000000-0000-0000-0000-000000000001", EmpresaID: "1", TerminalID: "2", Payload: `{"total":"10"}`}
	one, _ := s.Apply(op)
	two, _ := s.Apply(op)
	if one != two {
		t.Fatal("retry did not return original result")
	}
	if _, err := s.Apply(Operation{ID: op.ID, EmpresaID: op.EmpresaID, TerminalID: op.TerminalID, Payload: `{"total":"11"}`}); err != ErrConflict {
		t.Fatal("changed payload was accepted")
	}
	if _, err := s.Apply(Operation{ID: op.ID, EmpresaID: "2", TerminalID: op.TerminalID, Payload: op.Payload}); err != ErrConflict {
		t.Fatal("changed company scope was accepted")
	}
}

func TestEquivalentJSONFormattingIsIdempotent(t *testing.T) {
	s := NewStore()
	op := Operation{ID: "00000000-0000-0000-0000-000000000005", EmpresaID: "1", TerminalID: "2", Payload: `{ "total": "10" }`}
	if _, err := s.Apply(op); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Apply(Operation{ID: op.ID, EmpresaID: op.EmpresaID, TerminalID: op.TerminalID, Payload: `{"total":"10"}`}); err != nil {
		t.Fatalf("equivalent JSON was not idempotent: %v", err)
	}
	if ContentHash(op.Payload) != ContentHash(`{"total":"10"}`) {
		t.Fatal("equivalent JSON produced different content hashes")
	}
}

func TestDecodeRejectsInvalidPayloadJSON(t *testing.T) {
	if _, err := Decode([]byte(`{"id":"op-1","empresa_id":"1","terminal_id":"2","payload":"not-json"}`)); err != ErrInvalidOperation {
		t.Fatalf("expected invalid payload rejection, got %v", err)
	}
}

func TestDecodeRejectsNonUUIDOperationID(t *testing.T) {
	if _, err := Decode([]byte(`{"id":"op-1","empresa_id":"1","terminal_id":"2","payload":"{}"}`)); err != ErrInvalidOperation {
		t.Fatalf("expected UUID validation, got %v", err)
	}
}

func TestStoreRejectsInvalidOperationBeforePersistence(t *testing.T) {
	s := NewStore()
	if _, err := s.Apply(Operation{ID: "00000000-0000-0000-0000-000000000004", EmpresaID: "1", TerminalID: "2", Payload: "not-json"}); err != ErrInvalidOperation {
		t.Fatalf("expected invalid operation, got %v", err)
	}
}

func TestValidateRejectsInvalidCompanyScope(t *testing.T) {
	op := Operation{ID: "00000000-0000-0000-0000-000000000006", EmpresaID: "empresa", TerminalID: "2", Payload: `{}`}
	if err := Validate(op); err != ErrInvalidOperation {
		t.Fatalf("expected invalid company scope, got %v", err)
	}
}

func TestNilStoreReturnsControlledError(t *testing.T) {
	var s *Store
	_, err := s.Apply(Operation{})
	if err != ErrInvalidOperation {
		t.Fatalf("expected controlled nil store error, got %v", err)
	}
}
