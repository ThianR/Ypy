package api

import (
	"errors"
	"strconv"
)

var ErrAmbiguousScaleCode = errors.New("ambiguous scale code")
var ErrInvalidScaleValue = errors.New("invalid scale value")

type ScaleRule struct {
	Code, Prefix              string
	WeightDigits, PriceDigits int
	Priority                  int
}
type ScaleResult struct {
	ItemCode        string
	Quantity, Price string
}

// V4ScaleRule representa la regla posicional almacenada en gp_regla_balanza.
// Las posiciones comienzan en uno, según la definición de la tabla V4.
type V4ScaleRule struct {
	Prefix        string `json:"prefix"`
	Length        int    `json:"length"`
	ProductStart  int    `json:"product_start"`
	ProductLength int    `json:"product_length"`
	ValueStart    int    `json:"value_start"`
	ValueLength   int    `json:"value_length"`
	Decimals      int    `json:"decimals"`
	Content       string `json:"content"`
}

func (r V4ScaleRule) Valid() bool {
	return r.Prefix != "" && r.Length > 0 && len(r.Prefix) <= r.Length && r.ProductStart > 0 && r.ProductLength > 0 && r.ValueStart > 0 && r.ValueLength > 0 && r.Decimals >= 0 && r.Decimals < r.ValueLength && r.Decimals <= 6 && (r.Content == "PESO" || r.Content == "PRECIO") && r.ProductStart+r.ProductLength-1 <= r.Length && r.ValueStart+r.ValueLength-1 <= r.Length
}

func ParseV4ScaleCode(code string, rules []V4ScaleRule) (ScaleResult, error) {
	var selected *V4ScaleRule
	for i := range rules {
		r := &rules[i]
		if !r.Valid() || len(code) != r.Length || code[:len(r.Prefix)] != r.Prefix {
			continue
		}
		if selected != nil {
			return ScaleResult{}, ErrAmbiguousScaleCode
		}
		selected = r
	}
	if selected == nil {
		return ScaleResult{}, ErrInvalidScaleValue
	}
	productStart, valueStart := selected.ProductStart-1, selected.ValueStart-1
	product := code[productStart : productStart+selected.ProductLength]
	value := code[valueStart : valueStart+selected.ValueLength]
	if !digitsOnly(product) || !digitsOnly(value) {
		return ScaleResult{}, ErrInvalidScaleValue
	}
	amount, err := strconv.ParseInt(value, 10, 64)
	if err != nil || amount <= 0 {
		return ScaleResult{}, ErrInvalidScaleValue
	}
	return ScaleResult{ItemCode: product, Quantity: value, Price: value}, nil
}

func digitsOnly(value string) bool {
	if value == "" {
		return false
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}

func ParseScaleCode(code string, rules []ScaleRule) (ScaleResult, error) {
	var chosen *ScaleRule
	for i := range rules {
		r := &rules[i]
		if r.WeightDigits <= 0 || r.PriceDigits <= 0 || r.Code == "" {
			continue
		}
		if r.Code == code {
			chosen = r
			break
		}
	}
	if chosen == nil {
		for i := range rules {
			r := &rules[i]
			if r.WeightDigits <= 0 || r.PriceDigits <= 0 || r.Code == "" {
				continue
			}
			if r.Prefix != "" && len(code) >= len(r.Prefix) && code[:len(r.Prefix)] == r.Prefix {
				if chosen != nil && chosen.Priority == r.Priority {
					return ScaleResult{}, ErrAmbiguousScaleCode
				}
				if chosen == nil || r.Priority > chosen.Priority {
					chosen = r
				}
			}
		}
	}
	if chosen == nil || len(code) < len(chosen.Prefix)+chosen.WeightDigits+chosen.PriceDigits {
		return ScaleResult{}, ErrInvalidScaleValue
	}
	start := len(chosen.Prefix)
	weight, err1 := strconv.ParseInt(code[start:start+chosen.WeightDigits], 10, 64)
	price, err2 := strconv.ParseInt(code[start+chosen.WeightDigits:start+chosen.WeightDigits+chosen.PriceDigits], 10, 64)
	if err1 != nil || err2 != nil || weight <= 0 || price <= 0 {
		return ScaleResult{}, ErrInvalidScaleValue
	}
	// Conserva la representación de ancho fijo. El contrato POS utiliza los
	// dígitos codificados sin cambios para igualar los resultados de Go y TypeScript.
	return ScaleResult{ItemCode: chosen.Code, Quantity: code[start : start+chosen.WeightDigits], Price: code[start+chosen.WeightDigits : start+chosen.WeightDigits+chosen.PriceDigits]}, nil
}
