package api

import "testing"

func TestCashOpenCollectClose(t *testing.T) {
	s, err := Open("user-1", "cash-1", true, "100000")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Collect(Payment{Method: "CASH", Amount: "150000"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Collect(Payment{Method: "CARD", Amount: "50000"}); err != nil {
		t.Fatal(err)
	}
	difference, err := s.Close("290000")
	if err != nil || difference != "-10000.000000" {
		t.Fatalf("difference=%s err=%v", difference, err)
	}
}

func TestUnauthorizedOpen(t *testing.T) {
	if _, err := Open("user-1", "cash-1", false, "1"); err != ErrUnauthorizedOpen {
		t.Fatal("unauthorized opening accepted")
	}
}

func TestCashCloseCurrencies(t *testing.T) {
	s, err := Open("u", "c", true, "1")
	if err != nil {
		t.Fatal(err)
	}
	differences, err := s.CloseCurrencies(map[string]string{"PYG": "1000", "USD": "10.50"}, map[string]string{"PYG": "900", "USD": "10.75"})
	if err != nil || differences["PYG"] != "-100.000000" || differences["USD"] != "0.250000" || s.Open {
		t.Fatalf("unexpected multcurrency close: %#v %v", differences, err)
	}
}

func TestCashRejectsNegativeAmountsAndIncompleteIdentity(t *testing.T) {
	if _, err := Open("", "cash-1", true, "1"); err != ErrUnauthorizedOpen {
		t.Fatalf("expected incomplete identity rejection, got %v", err)
	}
	if _, err := Open("u", "c", true, "-1"); err != ErrInvalidCashAmount {
		t.Fatalf("expected negative opening rejection, got %v", err)
	}
	s, err := Open("u", "c", true, "1")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Collect(Payment{Method: "CASH", Amount: "-1"}); err != ErrInvalidCashAmount {
		t.Fatalf("expected negative payment rejection, got %v", err)
	}
}

func TestCashRejectsUnexpectedCurrencyAndNormalizesMethod(t *testing.T) {
	s, err := Open("u", "c", true, "0")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Collect(Payment{Method: " card ", Amount: "1"}); err != nil || s.Payments[0].Method != "CARD" {
		t.Fatalf("payment method was not normalized: %#v %v", s.Payments, err)
	}
	if _, err := s.CloseCurrencies(map[string]string{"PYG": "1"}, map[string]string{"PYG": "1", "USD": "1"}); err != ErrInvalidCashAmount {
		t.Fatalf("unexpected currency accepted: %v", err)
	}
}

func TestCashCloseCurrenciesRejectsEmptyExpectedSet(t *testing.T) {
	s, err := Open("u", "c", true, "0")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CloseCurrencies(map[string]string{}, map[string]string{}); err != ErrInvalidCashAmount || !s.Open {
		t.Fatalf("empty currency close accepted: %v open=%v", err, s.Open)
	}
}
