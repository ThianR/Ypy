package events

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidEnvelope = errors.New("invalid event envelope")

type Envelope struct {
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	EmpresaID     string          `json:"empresa_id"`
	Version       int             `json:"version"`
	Origin        string          `json:"origin"`
	CorrelationID string          `json:"correlation_id"`
	OccurredAt    string          `json:"occurred_at"`
	Payload       json.RawMessage `json:"payload"`
}

func (e Envelope) Validate() error {
	if _, err := uuid.Parse(e.EventID); err != nil {
		return ErrInvalidEnvelope
	}
	if _, err := uuid.Parse(e.CorrelationID); err != nil || strings.TrimSpace(e.EventType) == "" || strings.TrimSpace(e.Origin) == "" {
		return ErrInvalidEnvelope
	}
	company, err := strconv.ParseInt(e.EmpresaID, 10, 64)
	if err != nil || company <= 0 || e.Version < 1 || !json.Valid(e.Payload) || e.Payload[0] != '{' {
		return ErrInvalidEnvelope
	}
	if _, err := time.Parse(time.RFC3339, e.OccurredAt); err != nil {
		return ErrInvalidEnvelope
	}
	return nil
}
