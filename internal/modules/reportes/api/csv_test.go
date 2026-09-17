package api

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

func TestCSVExportHasStableColumns(t *testing.T) {
	csv := ToCSV([]Total{{Currency: "PYG", Method: "CASH", Amount: "150.000000"}})
	lines := strings.Split(strings.TrimSpace(csv), "\n")
	if lines[0] != "currency,method,amount" || !strings.Contains(lines[1], "PYG,CASH,150.000000") {
		t.Fatalf("unexpected CSV: %s", csv)
	}
}

func TestAggregateSalesIsDeterministic(t *testing.T) {
	got := AggregateSales([]Sale{
		{Currency: "USD", Method: "CARD", Total: "2"},
		{Currency: "PYG", Method: "CASH", Total: "100"},
		{Currency: "PYG", Method: "CARD", Total: "50"},
	})
	if got[0].Currency != "PYG" || got[0].Method != "CARD" || got[1].Method != "CASH" || got[2].Currency != "USD" {
		t.Fatalf("unexpected deterministic order: %#v", got)
	}
}

func TestOperationalCSVsRequireCompanyAndEscapeFields(t *testing.T) {
	stock := ToStockCSV([]StockRow{{CompanyID: 7, Item: "A,1", Location: "DEP", Reason: "RECEPCION", Quantity: "2", Value: "10.000000"}})
	if !strings.Contains(stock, `7,"A,1",DEP,RECEPCION,2,10.000000`) {
		t.Fatalf("unexpected stock csv: %s", stock)
	}
	if ToStockCSV([]StockRow{{Item: "A", Quantity: "1"}}) != "" {
		t.Fatal("unscoped stock row accepted")
	}
	cash := ToCashCSV([]CashRow{{CompanyID: 7, Currency: "PYG", Method: "CASH", Kind: "INGRESO", Amount: "100"}})
	if !strings.Contains(cash, "7,PYG,CASH,INGRESO,100") {
		t.Fatalf("unexpected cash csv: %s", cash)
	}
}

func TestTotalsXLSXIsValidAndStable(t *testing.T) {
	a, err := TotalsXLSX([]Total{{Currency: "USD", Method: "CARD", Amount: "2"}, {Currency: "PYG", Method: "CASH", Amount: "10"}})
	if err != nil {
		t.Fatal(err)
	}
	b, err := TotalsXLSX([]Total{{Currency: "PYG", Method: "CASH", Amount: "10"}, {Currency: "USD", Method: "CARD", Amount: "2"}})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("xlsx output is not deterministic")
	}
	z, err := zip.NewReader(bytes.NewReader(a), int64(len(a)))
	if err != nil {
		t.Fatal(err)
	}
	seen := false
	for _, f := range z.File {
		if f.Name == "xl/worksheets/sheet1.xml" {
			seen = true
		}
	}
	if !seen {
		t.Fatal("worksheet missing")
	}
}

func TestTotalsPDFIsValidAndStable(t *testing.T) {
	a, err := TotalsPDF([]Total{{Currency: "USD", Method: "CARD", Amount: "2"}, {Currency: "PYG", Method: "CASH", Amount: "10"}})
	if err != nil {
		t.Fatal(err)
	}
	b, err := TotalsPDF([]Total{{Currency: "PYG", Method: "CASH", Amount: "10"}, {Currency: "USD", Method: "CARD", Amount: "2"}})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) || !bytes.HasPrefix(a, []byte("%PDF-1.4")) {
		t.Fatal("invalid or nondeterministic pdf")
	}
}
