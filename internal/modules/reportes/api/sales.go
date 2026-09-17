package api

import (
	"encoding/csv"
	"errors"
	"math/big"
	"regexp"
	"sort"
	"strings"
)

type Sale struct{ Currency, Method, Total string }
type Total struct{ Currency, Method, Amount string }

var ErrInvalidSale = errors.New("invalid report sale")
var decimalAmount = regexp.MustCompile(`^(0|[1-9][0-9]*)(\.[0-9]+)?$`)

func AggregateSales(sales []Sale) []Total {
	result, err := AggregateSalesValidated(sales)
	if err != nil {
		return nil
	}
	return result
}

func AggregateSalesValidated(sales []Sale) ([]Total, error) {
	sums := map[string]*big.Rat{}
	for _, sale := range sales {
		currency := strings.TrimSpace(sale.Currency)
		method := strings.TrimSpace(sale.Method)
		if currency == "" || method == "" {
			return nil, ErrInvalidSale
		}
		key := currency + "/" + method
		value, ok := new(big.Rat).SetString(sale.Total)
		if !decimalAmount.MatchString(sale.Total) || !ok || value.Sign() < 0 {
			return nil, ErrInvalidSale
		}
		if sums[key] == nil {
			sums[key] = new(big.Rat)
		}
		sums[key].Add(sums[key], value)
	}
	result := make([]Total, 0, len(sums))
	for key, value := range sums {
		var currency, method string
		for i, r := range key {
			if r == '/' {
				currency, method = key[:i], key[i+1:]
				break
			}
		}
		result = append(result, Total{Currency: currency, Method: method, Amount: value.FloatString(6)})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Currency == result[j].Currency {
			return result[i].Method < result[j].Method
		}
		return result[i].Currency < result[j].Currency
	})
	return result, nil
}

func ToCSV(totals []Total) string {
	return ToCSVValidated(totals)
}

func ToCSVValidated(totals []Total) string {
	clean := make([]Total, 0, len(totals))
	for _, total := range totals {
		if strings.TrimSpace(total.Currency) == "" || strings.TrimSpace(total.Method) == "" || !decimalAmount.MatchString(total.Amount) {
			return ""
		}
		clean = append(clean, Total{Currency: strings.TrimSpace(total.Currency), Method: strings.TrimSpace(total.Method), Amount: total.Amount})
	}
	sort.Slice(clean, func(i, j int) bool {
		if clean[i].Currency == clean[j].Currency {
			return clean[i].Method < clean[j].Method
		}
		return clean[i].Currency < clean[j].Currency
	})
	var b strings.Builder
	w := csv.NewWriter(&b)
	_ = w.Write([]string{"currency", "method", "amount"})
	for _, t := range clean {
		_ = w.Write([]string{t.Currency, t.Method, t.Amount})
	}
	w.Flush()
	return b.String()
}
