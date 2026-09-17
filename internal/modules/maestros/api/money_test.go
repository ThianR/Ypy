package api

import "testing"

func TestAddDecimalExact(t *testing.T) {
	got, err := AddDecimal("0.1", "0.2")
	if err != nil || got != "0.300000" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestRejectsFloatLikeInvalidInput(t *testing.T) {
	if _, err := AddDecimal("not-a-number"); err == nil {
		t.Fatal("invalid amount accepted")
	}
}
