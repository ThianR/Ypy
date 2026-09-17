package api

import "errors"

var ErrInsufficientOrigin = errors.New("insufficient origin stock")
var ErrDuplicateTransfer = errors.New("transfer already applied")
var ErrTransferConflict = errors.New("transfer id reused with different content")
var ErrInvalidTransfer = errors.New("invalid transfer")

type Transfer struct {
	ID, ItemID, Origin, Destination string
	Quantity                        int64
}
type Stock struct {
	balances map[string]int64
	applied  map[string]Transfer
}

func NewStock() *Stock { return &Stock{balances: map[string]int64{}, applied: map[string]Transfer{}} }
func (s *Stock) Set(location, item string, quantity int64) {
	if s != nil && location != "" && item != "" && quantity >= 0 {
		s.balances[location+"/"+item] = quantity
	}
}
func (s *Stock) Transfer(t Transfer) error {
	if s == nil || t.ID == "" || t.ItemID == "" || t.Origin == "" || t.Destination == "" || t.Origin == t.Destination || t.Quantity <= 0 {
		return ErrInvalidTransfer
	}
	if previous, ok := s.applied[t.ID]; ok {
		if previous.ItemID != t.ItemID || previous.Origin != t.Origin || previous.Destination != t.Destination || previous.Quantity != t.Quantity {
			return ErrTransferConflict
		}
		return ErrDuplicateTransfer
	}
	keyOrigin, keyDestination := t.Origin+"/"+t.ItemID, t.Destination+"/"+t.ItemID
	if s.balances[keyOrigin] < t.Quantity {
		return ErrInsufficientOrigin
	}
	s.balances[keyOrigin] -= t.Quantity
	s.balances[keyDestination] += t.Quantity
	s.applied[t.ID] = t
	return nil
}
func (s *Stock) Balance(location, item string) int64 { return s.balances[location+"/"+item] }
