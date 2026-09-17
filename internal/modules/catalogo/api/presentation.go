package api

import (
	"errors"
	"math/big"
)

var ErrInvalidConversion = errors.New("invalid presentation conversion")

func ToBaseQuantity(quantity, factor string) (string, error) {
	q, ok := new(big.Rat).SetString(quantity)
	if !ok || q.Sign() < 0 {
		return "", ErrInvalidConversion
	}
	f, ok := new(big.Rat).SetString(factor)
	if !ok || f.Sign() <= 0 {
		return "", ErrInvalidConversion
	}
	return new(big.Rat).Mul(q, f).FloatString(6), nil
}
