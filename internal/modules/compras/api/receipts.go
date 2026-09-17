package api

import "errors"

var ErrOverReceipt = errors.New("receipt exceeds ordered quantity")
var ErrDuplicateReceipt = errors.New("receipt already registered")
var ErrReceiptConflict = errors.New("receipt id reused with different quantity")
var ErrInvalidOrder = errors.New("invalid purchase order")

type Order struct {
	ID, ItemID        string
	Ordered, Received int64
	receipts          map[string]int64
}

func NewOrder(id, itemID string, quantity int64) *Order {
	return &Order{ID: id, ItemID: itemID, Ordered: quantity, receipts: map[string]int64{}}
}
func (o *Order) Receive(receiptID string, quantity int64) error {
	if o == nil || o.ID == "" || o.ItemID == "" || o.Ordered <= 0 || receiptID == "" {
		return ErrInvalidOrder
	}
	if previous, ok := o.receipts[receiptID]; ok {
		if previous != quantity {
			return ErrReceiptConflict
		}
		return ErrDuplicateReceipt
	}
	if quantity <= 0 || o.Received+quantity > o.Ordered {
		return ErrOverReceipt
	}
	o.receipts[receiptID] = quantity
	o.Received += quantity
	return nil
}
func (o *Order) Pending() int64 { return o.Ordered - o.Received }
