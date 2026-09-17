package workflows

import (
	"errors"
	"fmt"
	"github.com/ypy-erp/ypy/internal/modules/inventario/api"
	numeracion "github.com/ypy-erp/ypy/internal/modules/numeracion/api"
	"math/big"
)

var ErrPaymentFailed = errors.New("payment failed")
var ErrInvalidSale = errors.New("invalid sale command")
var ErrSaleConflict = errors.New("sale id reused with different content")

type Sale struct {
	ID, ItemID, Number string
	Quantity           int64
	Total              string
	Confirmed          bool
}
type SaleCommand struct {
	ID, TerminalID, ItemID, Total string
	Quantity                      int64
	PaymentOK                     bool
}
type Sales struct {
	nextNumber int64
	stock      *api.Ledger
	applied    map[string]Sale
}

func NewSales(firstNumber int64, stock *api.Ledger) *Sales {
	return &Sales{nextNumber: firstNumber, stock: stock, applied: map[string]Sale{}}
}

func (s *Sales) Confirm(c SaleCommand) (Sale, error) {
	if s == nil {
		return Sale{}, ErrInvalidSale
	}
	if previous, ok := s.applied[c.ID]; ok {
		if previous.ItemID != c.ItemID || previous.Quantity != c.Quantity || previous.Total != c.Total {
			return Sale{}, ErrSaleConflict
		}
		return previous, nil
	}
	if !c.PaymentOK {
		return Sale{}, ErrPaymentFailed
	}
	total, validTotal := new(big.Rat).SetString(c.Total)
	if s.stock == nil || c.ID == "" || c.TerminalID == "" || c.ItemID == "" || c.Quantity <= 0 || !validTotal || total.Sign() < 0 {
		return Sale{}, ErrInvalidSale
	}
	block := numeracion.NewAllocator(s.nextNumber).Reserve(c.TerminalID, 1)
	if block == nil {
		return Sale{}, numeracion.ErrInvalidRange
	}
	n, err := block.Consume(c.TerminalID)
	if err != nil {
		return Sale{}, err
	}
	if err := s.stock.Issue(c.ID, c.ItemID, c.Quantity); err != nil {
		return Sale{}, err
	}
	s.nextNumber = n + 1
	result := Sale{ID: c.ID, ItemID: c.ItemID, Number: formatNumber(n), Quantity: c.Quantity, Total: c.Total, Confirmed: true}
	s.applied[c.ID] = result
	return result, nil
}

func formatNumber(n int64) string { return fmt.Sprintf("%d", n) }
