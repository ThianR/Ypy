package api

import (
	"errors"
	"strings"
)

var ErrMovementReasonRequired = errors.New("movement reason required")
var ErrDuplicateMovement = errors.New("movement already registered")
var ErrMovementConflict = errors.New("movement id reused with different content")
var ErrInvalidMovement = errors.New("invalid cash movement")

type Movement struct {
	ID, SessionID, Kind, Reason string
	Amount                      int64
}
type MovementBook struct{ entries map[string]Movement }

func NewMovementBook() *MovementBook { return &MovementBook{entries: map[string]Movement{}} }
func (b *MovementBook) Add(m Movement) error {
	if b == nil || m.ID == "" || m.SessionID == "" || (m.Kind != "INCOME" && m.Kind != "WITHDRAWAL") || m.Amount <= 0 {
		return ErrInvalidMovement
	}
	if previous, ok := b.entries[m.ID]; ok {
		if previous.SessionID != m.SessionID || previous.Kind != m.Kind || previous.Reason != strings.TrimSpace(m.Reason) || previous.Amount != m.Amount {
			return ErrMovementConflict
		}
		return ErrDuplicateMovement
	}
	m.Reason = strings.TrimSpace(m.Reason)
	if m.Reason == "" {
		return ErrMovementReasonRequired
	}
	b.entries[m.ID] = m
	return nil
}
func (b *MovementBook) Count() int {
	if b == nil {
		return 0
	}
	return len(b.entries)
}
