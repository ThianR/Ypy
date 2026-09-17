package api

import (
	"errors"
	"math/big"
	"strings"
)

var ErrUnauthorizedOpen = errors.New("user is not authorized to open cash")
var ErrCashClosed = errors.New("cash session is closed")
var ErrInvalidCashAmount = errors.New("invalid cash amount")

type Payment struct{ Method, Amount string }
type Session struct {
	UserID, CashID string
	Open           bool
	Expected       string
	Payments       []Payment
}

func Open(userID, cashID string, authorized bool, opening string) (*Session, error) {
	if !authorized || userID == "" || cashID == "" {
		return nil, ErrUnauthorizedOpen
	}
	if !nonNegative(opening) {
		return nil, ErrInvalidCashAmount
	}
	return &Session{UserID: userID, CashID: cashID, Open: true, Expected: opening}, nil
}

func (s *Session) Collect(p Payment) error {
	if s == nil || !s.Open {
		return ErrCashClosed
	}
	p.Method = strings.ToUpper(strings.TrimSpace(p.Method))
	if !nonNegative(p.Amount) || p.Method == "" {
		return ErrInvalidCashAmount
	}
	amount, _ := new(big.Rat).SetString(p.Amount)
	expected, _ := new(big.Rat).SetString(s.Expected)
	expected.Add(expected, amount)
	s.Expected = expected.FloatString(6)
	s.Payments = append(s.Payments, p)
	return nil
}

func (s *Session) Close(counted string) (string, error) {
	if s == nil || !s.Open {
		return "", ErrCashClosed
	}
	if !nonNegative(counted) {
		return "", ErrInvalidCashAmount
	}
	actual, _ := new(big.Rat).SetString(counted)
	expected, _ := new(big.Rat).SetString(s.Expected)
	s.Open = false
	return new(big.Rat).Sub(actual, expected).FloatString(6), nil
}

func (s *Session) CloseCurrencies(expected, counted map[string]string) (map[string]string, error) {
	if s == nil || !s.Open {
		return nil, ErrCashClosed
	}
	if len(expected) == 0 {
		return nil, ErrInvalidCashAmount
	}
	differences := make(map[string]string, len(expected))
	for currency, value := range expected {
		if !nonNegative(value) {
			return nil, ErrInvalidCashAmount
		}
		if !nonNegative(counted[currency]) {
			return nil, ErrInvalidCashAmount
		}
		expectedValue, _ := new(big.Rat).SetString(value)
		countedValue, _ := new(big.Rat).SetString(counted[currency])
		differences[currency] = new(big.Rat).Sub(countedValue, expectedValue).FloatString(6)
	}
	for currency, value := range counted {
		if _, exists := expected[currency]; !exists || !nonNegative(value) {
			return nil, ErrInvalidCashAmount
		}
	}
	s.Open = false
	return differences, nil
}

func nonNegative(value string) bool {
	n, ok := new(big.Rat).SetString(value)
	return ok && n.Sign() >= 0
}
