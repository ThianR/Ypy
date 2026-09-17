package poscentral

import "testing"

func TestDecodeRejectsIncompleteOperation(t *testing.T) {
	if _, err := Decode([]byte(`{"id":"op-1","empresa_id":"1"}`)); err != ErrInvalidOperation {
		t.Fatal("incomplete operation accepted")
	}
}
