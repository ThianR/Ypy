package api

import (
	"errors"
	"math/big"
	"regexp"
)

var ErrMissingRate = errors.New("exchange rate unavailable")
var ErrCurrencyMismatch = errors.New("payment currency does not match quote")
var decimalAmount = regexp.MustCompile(`^(0|[1-9][0-9]*)(\.[0-9]+)?$`)

type Quote struct{ From, To, Rate, AsOf string }
type Payment struct{ Amount, Currency string }

func SumPaymentsInBase(payments []Payment, base string, quotes map[string]Quote) (string, error) {
	if base == "" || len(payments) == 0 {
		return "", ErrMissingRate
	}
	total := new(big.Rat)
	for _, payment := range payments {
		if payment.Currency == "" {
			return "", ErrCurrencyMismatch
		}
		amount, ok := new(big.Rat).SetString(payment.Amount)
		if !decimalAmount.MatchString(payment.Amount) || !ok || amount.Sign() < 0 {
			return "", ErrInvalidAmount
		}
		if payment.Currency == base {
			total.Add(total, amount)
			continue
		}
		quote, ok := quotes[payment.Currency]
		if !ok || quote.To != base || quote.From != payment.Currency {
			return "", ErrMissingRate
		}
		converted, err := Convert(payment.Amount, quote)
		if err != nil {
			return "", err
		}
		value, _ := new(big.Rat).SetString(converted)
		total.Add(total, value)
	}
	return total.FloatString(6), nil
}

func Convert(amount string, quote Quote) (string, error) {
	if quote.From == "" || quote.To == "" || quote.AsOf == "" || quote.From == quote.To {
		return "", ErrMissingRate
	}
	a, ok := new(big.Rat).SetString(amount)
	if !decimalAmount.MatchString(amount) || !ok || a.Sign() < 0 {
		return "", ErrInvalidAmount
	}
	r, ok := new(big.Rat).SetString(quote.Rate)
	if !ok || r.Sign() <= 0 {
		return "", ErrMissingRate
	}
	return new(big.Rat).Mul(a, r).FloatString(6), nil
}
