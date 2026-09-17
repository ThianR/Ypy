package api

import "testing"

func TestSerialCannotBeRegisteredTwice(t *testing.T) {
	i := NewTrackedItem("item", true)
	if err := i.RegisterSerial("S-1"); err != nil {
		t.Fatal(err)
	}
	if err := i.RegisterSerial("S-1"); err != ErrDuplicateSerial {
		t.Fatal("duplicate serial accepted")
	}
}
func TestRequiredTraceabilityMustBePresent(t *testing.T) {
	i := NewTrackedItem("item", true)
	if err := i.ValidateSale(nil); err != ErrTraceabilityRequired {
		t.Fatal("sale without serial accepted")
	}
}

func TestUnknownSerialIsNotReportedAsDuplicate(t *testing.T) {
	i := NewTrackedItem("item", true)
	if err := i.ValidateSale([]string{"missing"}); err != ErrUnknownSerial {
		t.Fatalf("expected unknown serial, got %v", err)
	}
}

func TestEmptySerialIsRejected(t *testing.T) {
	i := NewTrackedItem("item", true)
	if err := i.RegisterSerial(""); err != ErrUnknownSerial {
		t.Fatalf("expected empty serial rejection, got %v", err)
	}
}

func TestSaleRejectsRepeatedSerialAndNormalizesInput(t *testing.T) {
	i := NewTrackedItem("item", true)
	if err := i.RegisterSerial(" S-1 "); err != nil {
		t.Fatal(err)
	}
	if err := i.ValidateSale([]string{"S-1", " S-1 "}); err != ErrDuplicateSerial {
		t.Fatalf("repeated serial accepted: %v", err)
	}
}
