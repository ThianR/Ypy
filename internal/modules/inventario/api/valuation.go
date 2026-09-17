package api

import (
	"errors"
	"math/big"
)

var ErrInvalidCost = errors.New("invalid cost")
var ErrCostAdjustmentRequired = errors.New("cost adjustment requires pending state and reason")

type Valuation struct {
	Quantity, Cost string
	PendingCost    bool
}

func (v Valuation) Receive(quantity, unitCost string) (Valuation, error) {
	q, ok := new(big.Rat).SetString(quantity)
	if !ok || q.Sign() <= 0 {
		return v, ErrInvalidCost
	}
	c, ok := new(big.Rat).SetString(unitCost)
	if !ok || c.Sign() < 0 {
		return v, ErrInvalidCost
	}
	oldQ, qOK := new(big.Rat).SetString(v.Quantity)
	oldC, cOK := new(big.Rat).SetString(v.Cost)
	if !qOK || !cOK || oldQ.Sign() < 0 || oldC.Sign() < 0 {
		return v, ErrInvalidCost
	}
	total := new(big.Rat).Add(new(big.Rat).Mul(oldQ, oldC), new(big.Rat).Mul(q, c))
	newQ := new(big.Rat).Add(oldQ, q)
	return Valuation{Quantity: newQ.FloatString(6), Cost: new(big.Rat).Quo(total, newQ).FloatString(6)}, nil
}

func (v Valuation) MarkCostPending() Valuation { v.PendingCost = true; return v }

// ResolvePendingCost hace explícito el ajuste autorizado del costo pendiente.
func (v Valuation) ResolvePendingCost(unitCost, reason string) (Valuation, error) {
	if !v.PendingCost || reason == "" {
		return v, ErrCostAdjustmentRequired
	}
	c, ok := new(big.Rat).SetString(unitCost)
	if !ok || c.Sign() < 0 {
		return v, ErrInvalidCost
	}
	if _, ok := new(big.Rat).SetString(v.Quantity); !ok {
		return v, ErrInvalidCost
	}
	v.Cost = c.FloatString(6)
	v.PendingCost = false
	return v, nil
}
