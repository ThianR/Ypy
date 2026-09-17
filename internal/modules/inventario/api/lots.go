package api

import (
	"errors"
	"time"
)

var ErrExpiredLot = errors.New("lot expired")

type Lot struct {
	ID, ItemID string
	ExpiresAt  time.Time
}

func (l Lot) CanSell(at time.Time) error {
	if l.ID == "" || l.ItemID == "" {
		return ErrExpiredLot
	}
	if !l.ExpiresAt.IsZero() && !at.Before(l.ExpiresAt) {
		return ErrExpiredLot
	}
	return nil
}
