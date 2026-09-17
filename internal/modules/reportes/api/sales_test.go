package api

import "testing"

func TestAggregateKeepsCurrencyAndMethod(t *testing.T) {
	got := AggregateSales([]Sale{{"PYG", "CASH", "100"}, {"PYG", "CASH", "50"}, {"USD", "CARD", "2"}})
	if len(got) != 2 {
		t.Fatalf("groups=%d", len(got))
	}
	found := false
	for _, total := range got {
		if total.Currency == "PYG" && total.Method == "CASH" {
			found = total.Amount == "150.000000"
		}
	}
	if !found {
		t.Fatal("PYG cash total did not reconcile")
	}
}

func TestAggregateNormalizesReportDimensions(t *testing.T) {
	got, err := AggregateSalesValidated([]Sale{{Currency: " PYG ", Method: " CASH ", Total: "1"}, {Currency: "PYG", Method: "CASH", Total: "2"}})
	if err != nil || len(got) != 1 || got[0].Currency != "PYG" || got[0].Method != "CASH" || got[0].Amount != "3.000000" {
		t.Fatalf("unexpected normalized report: %#v err=%v", got, err)
	}
}

func TestAggregateValidatedRejectsMalformedSale(t *testing.T) {
	if _, err := AggregateSalesValidated([]Sale{{Currency: "PYG", Method: "CASH", Total: "not-money"}}); err != ErrInvalidSale {
		t.Fatalf("expected invalid sale, got %v", err)
	}
	if _, err := AggregateSalesValidated([]Sale{{Currency: "", Method: "CASH", Total: "1"}}); err != ErrInvalidSale {
		t.Fatalf("expected missing currency rejection, got %v", err)
	}
	if _, err := AggregateSalesValidated([]Sale{{Currency: "PYG", Method: "CASH", Total: "1/2"}}); err != ErrInvalidSale {
		t.Fatalf("expected fractional amount rejection, got %v", err)
	}
}

func TestCSVExportIsEscapedAndDeterministic(t *testing.T) {
	got := ToCSVValidated([]Total{{Currency: "USD", Method: "CARD, POS", Amount: "2.000000"}, {Currency: "PYG", Method: "CASH", Amount: "3.000000"}})
	expected := "currency,method,amount\nPYG,CASH,3.000000\nUSD,\"CARD, POS\",2.000000\n"
	if got != expected {
		t.Fatalf("unexpected csv: %q", got)
	}
	if ToCSVValidated([]Total{{Currency: "PYG", Method: "CASH", Amount: "1/2"}}) != "" {
		t.Fatal("invalid total exported")
	}
}
