package api

import "errors"

var ErrReturnExceedsSale = errors.New("return exceeds sold quantity")
var ErrDuplicateReturn = errors.New("return already registered")
var ErrReturnConflict = errors.New("return id reused with different content")
var ErrInvalidReturn = errors.New("invalid return")

type SaleLine struct {
	SaleID, ItemID string
	Sold, Returned int64
	returns        map[string]int64
}

func NewSaleLine(saleID, itemID string, quantity int64) *SaleLine {
	return &SaleLine{SaleID: saleID, ItemID: itemID, Sold: quantity, returns: map[string]int64{}}
}
func (s *SaleLine) Return(id string, quantity int64) error {
	if s == nil || s.SaleID == "" || s.ItemID == "" || s.Sold <= 0 || id == "" {
		return ErrInvalidReturn
	}
	if previous, ok := s.returns[id]; ok {
		if previous != quantity {
			return ErrReturnConflict
		}
		return ErrDuplicateReturn
	}
	if quantity <= 0 || s.Returned+quantity > s.Sold {
		return ErrReturnExceedsSale
	}
	s.returns[id] = quantity
	s.Returned += quantity
	return nil
}
func (s *SaleLine) Returnable() int64 {
	if s == nil {
		return 0
	}
	return s.Sold - s.Returned
}
