package workflows

import (
	"github.com/ypy-erp/ypy/internal/modules/inventario/api"
	numeracion "github.com/ypy-erp/ypy/internal/modules/numeracion/api"
	"testing"
)

func TestSaleConfirmAndRetryIsIdempotent(t *testing.T) {
	stock := api.NewLedger()
	_ = stock.Receive(api.Receipt{ID: "r1", ItemID: "item", Quantity: 10})
	sales := NewSales(1000, stock)
	cmd := SaleCommand{ID: "sale-1", TerminalID: "t1", ItemID: "item", Quantity: 2, Total: "3000", PaymentOK: true}
	one, err := sales.Confirm(cmd)
	if err != nil {
		t.Fatal(err)
	}
	two, err := sales.Confirm(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if one.Number != two.Number || stock.Balance("item") != 8 {
		t.Fatalf("duplicate sale effect: %#v balance=%d", two, stock.Balance("item"))
	}
}

func TestSaleRejectsSameIDWithDifferentContent(t *testing.T) {
	stock := api.NewLedger()
	_ = stock.Receive(api.Receipt{ID: "r1", ItemID: "item", Quantity: 10})
	sales := NewSales(1000, stock)
	cmd := SaleCommand{ID: "sale-1", TerminalID: "t1", ItemID: "item", Quantity: 2, Total: "3000", PaymentOK: true}
	if _, err := sales.Confirm(cmd); err != nil {
		t.Fatal(err)
	}
	cmd.Quantity = 3
	if _, err := sales.Confirm(cmd); err != ErrSaleConflict {
		t.Fatalf("expected content conflict, got %v", err)
	}
	if stock.Balance("item") != 8 {
		t.Fatalf("conflict changed stock: %d", stock.Balance("item"))
	}
}

func TestFailedPaymentDoesNotChangeStock(t *testing.T) {
	stock := api.NewLedger()
	_ = stock.Receive(api.Receipt{ID: "r1", ItemID: "item", Quantity: 10})
	_, err := NewSales(1000, stock).Confirm(SaleCommand{ID: "sale-1", ItemID: "item", Quantity: 2, PaymentOK: false})
	if err != ErrPaymentFailed || stock.Balance("item") != 10 {
		t.Fatal("failed payment changed stock")
	}
}

func TestInvalidSaleConfigurationDoesNotPanic(t *testing.T) {
	stock := api.NewLedger()
	_, err := NewSales(0, stock).Confirm(SaleCommand{ID: "sale-1", TerminalID: "t1", ItemID: "item", Quantity: 1, Total: "10", PaymentOK: true})
	if err != numeracion.ErrInvalidRange {
		t.Fatalf("expected invalid range, got %v", err)
	}
}

func TestInvalidSaleCommandDoesNotReserveOrChangeStock(t *testing.T) {
	stock := api.NewLedger()
	_ = stock.Receive(api.Receipt{ID: "r1", ItemID: "item", Quantity: 10})
	_, err := NewSales(1000, stock).Confirm(SaleCommand{ID: "sale-1", TerminalID: "t1", ItemID: "item", Quantity: 0, Total: "10", PaymentOK: true})
	if err != ErrInvalidSale || stock.Balance("item") != 10 {
		t.Fatalf("invalid sale changed state: %v balance=%d", err, stock.Balance("item"))
	}
}

func TestNilSalesServiceReturnsControlledError(t *testing.T) {
	var sales *Sales
	if _, err := sales.Confirm(SaleCommand{PaymentOK: true}); err != ErrInvalidSale {
		t.Fatalf("expected invalid service, got %v", err)
	}
}

func TestSaleRejectsInvalidTotal(t *testing.T) {
	stock := api.NewLedger()
	for _, total := range []string{"bad", "-10"} {
		_, err := NewSales(1000, stock).Confirm(SaleCommand{ID: "sale-" + total, TerminalID: "t1", ItemID: "item", Quantity: 1, Total: total, PaymentOK: true})
		if err != ErrInvalidSale {
			t.Fatalf("expected invalid total %q rejection, got %v", total, err)
		}
	}
}
