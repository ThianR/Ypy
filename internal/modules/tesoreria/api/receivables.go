package api

import "errors"

var ErrOverpayment = errors.New("payment exceeds pending balance")
var ErrDuplicateApplication = errors.New("payment application already registered")
var ErrApplicationConflict = errors.New("payment id reused with different amount")
var ErrInvalidReceivable = errors.New("invalid receivable")

type Receivable struct {
	ID, Currency   string
	Total, Applied int64
	applications   map[string]int64
}

func NewReceivable(id, currency string, total int64) *Receivable {
	return &Receivable{ID: id, Currency: currency, Total: total, applications: map[string]int64{}}
}
func (r *Receivable) Apply(paymentID string, amount int64) error {
	if r == nil || r.ID == "" || r.Currency == "" || r.Total <= 0 || paymentID == "" {
		return ErrInvalidReceivable
	}
	if previous, ok := r.applications[paymentID]; ok {
		if previous != amount {
			return ErrApplicationConflict
		}
		return ErrDuplicateApplication
	}
	if amount <= 0 || r.Applied+amount > r.Total {
		return ErrOverpayment
	}
	r.applications[paymentID] = amount
	r.Applied += amount
	return nil
}
func (r *Receivable) Pending() int64 { return r.Total - r.Applied }
