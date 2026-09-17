package api

import "errors"

var ErrDuplicateReceipt = errors.New("receipt already applied")
var ErrReceiptConflict = errors.New("receipt id reused with different content")
var ErrInvalidReceipt = errors.New("invalid receipt")

type Receipt struct {
	ID, ItemID string
	Quantity   int64
}
type Ledger struct {
	applied map[string]Receipt
	balance map[string]int64
}

// Issue registra una salida de stock con una cantidad solicitada positiva.
func (l *Ledger) Issue(id, itemID string, quantity int64) error {
	if l == nil || id == "" || itemID == "" || quantity <= 0 {
		return ErrInvalidReceipt
	}
	if _, ok := l.applied[id]; ok {
		return ErrDuplicateReceipt
	}
	l.applied[id] = Receipt{ID: id, ItemID: itemID, Quantity: -quantity}
	l.balance[itemID] -= quantity
	return nil
}

func NewLedger() *Ledger { return &Ledger{applied: map[string]Receipt{}, balance: map[string]int64{}} }

// Receive aplica una recepción una sola vez, usando el ID documental como clave idempotente.
func (l *Ledger) Receive(r Receipt) error {
	if l == nil || r.ID == "" || r.ItemID == "" || r.Quantity <= 0 {
		return ErrInvalidReceipt
	}
	if previous, ok := l.applied[r.ID]; ok {
		if previous.ItemID != r.ItemID || previous.Quantity != r.Quantity {
			return ErrReceiptConflict
		}
		return ErrDuplicateReceipt
	}
	l.applied[r.ID] = r
	l.balance[r.ItemID] += r.Quantity
	return nil
}

func (l *Ledger) Balance(itemID string) int64 { return l.balance[itemID] }
