package api

import "testing"

func TestPresentationConvertsToBaseUnits(t *testing.T) {
	got, err := ToBaseQuantity("3", "12")
	if err != nil || got != "36.000000" {
		t.Fatalf("got=%s err=%v", got, err)
	}
}
func TestInvalidConversionRejected(t *testing.T) {
	if _, err := ToBaseQuantity("3", "0"); err != ErrInvalidConversion {
		t.Fatal("zero conversion accepted")
	}
}
