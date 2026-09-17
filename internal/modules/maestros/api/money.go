package api

import (
	"errors"
	"math/big"
)

var ErrInvalidAmount = errors.New("invalid decimal amount")

// AddDecimal suma decimales exactos transportados como cadenas.
func AddDecimal(values ...string) (string, error) {
	total := new(big.Rat)
	for _, value := range values {
		n, ok := new(big.Rat).SetString(value)
		if !ok {
			return "", ErrInvalidAmount
		}
		total.Add(total, n)
	}
	return total.FloatString(6), nil
}
