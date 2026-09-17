package api

import "testing"

func TestConvertUsesFixedQuote(t *testing.T) {
	q := Quote{From: "USD", To: "PYG", Rate: "7500.25", AsOf: "2026-09-16"}
	got, err := Convert("2", q)
	if err != nil || got != "15000.500000" {
		t.Fatalf("got %s, %v", got, err)
	}
}

func TestMissingQuoteRejected(t *testing.T) {
	if _, err := Convert("2", Quote{From: "USD", To: "PYG"}); err != ErrMissingRate {
		t.Fatal("missing quote accepted")
	}
}

func TestQuoteRequiresHistoricalSnapshotMetadata(t *testing.T) {
	if _, err := Convert("2", Quote{Rate: "7500"}); err != ErrMissingRate {
		t.Fatalf("expected incomplete quote rejection, got %v", err)
	}
	if _, err := Convert("2", Quote{From: "PYG", To: "PYG", Rate: "1", AsOf: "2026-09-16"}); err != ErrMissingRate {
		t.Fatalf("expected same-currency quote rejection, got %v", err)
	}
}

func TestConvertRejectsNegativeAmount(t *testing.T) {
	if _, err := Convert("-1", Quote{From: "USD", To: "PYG", Rate: "7000", AsOf: "2026-09-16"}); err != ErrInvalidAmount {
		t.Fatalf("negative amount accepted: %v", err)
	}
}

func TestSumMixedPaymentsUsesFixedQuotes(t *testing.T) {
	got, err := SumPaymentsInBase([]Payment{{Amount: "100", Currency: "PYG"}, {Amount: "2", Currency: "USD"}}, "PYG", map[string]Quote{"USD": {From: "USD", To: "PYG", Rate: "7500", AsOf: "2026-09-16"}})
	if err != nil || got != "15100.000000" {
		t.Fatalf("mixed payments=%s err=%v", got, err)
	}
	if _, err := SumPaymentsInBase([]Payment{{Amount: "2", Currency: "USD"}}, "PYG", nil); err != ErrMissingRate {
		t.Fatalf("missing quote accepted: %v", err)
	}
}

func TestConvertRejectsFractionSyntax(t *testing.T) {
	if _, err := Convert("1/2", Quote{From: "USD", To: "PYG", Rate: "7000", AsOf: "2026-09-16"}); err != ErrInvalidAmount {
		t.Fatalf("fraction syntax accepted: %v", err)
	}
}
