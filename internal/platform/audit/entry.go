package audit

import (
	"errors"
	"strings"
)

var ErrReasonRequired = errors.New("audit reason required")
var ErrInvalidEntry = errors.New("invalid audit entry")

type Entry struct {
	OperationID string
	Action      string
	Reason      string
	Rejected    bool
	Before      map[string]string
	After       map[string]string
}

// RejectedEntry registra un intento denegado sin representarlo como cambio aplicado.
func RejectedEntry(operationID, action, reason string, before map[string]string) (Entry, error) {
	entry, err := New(operationID, action, reason, before, nil)
	if err != nil {
		return Entry{}, err
	}
	entry.Rejected = true
	return entry, nil
}

func New(operationID, action, reason string, before, after map[string]string) (Entry, error) {
	if strings.TrimSpace(operationID) == "" || strings.TrimSpace(action) == "" {
		return Entry{}, ErrInvalidEntry
	}
	if strings.TrimSpace(reason) == "" && requiresReason(action) {
		return Entry{}, ErrReasonRequired
	}
	return Entry{OperationID: operationID, Action: action, Reason: strings.TrimSpace(reason), Before: Redact(before), After: Redact(after)}, nil
}

func requiresReason(action string) bool {
	action = strings.ToUpper(action)
	return strings.Contains(action, "ANULAR") || strings.Contains(action, "AJUST") || strings.Contains(action, "DEVOL") || strings.Contains(action, "OVERRIDE")
}

func Redact(values map[string]string) map[string]string {
	result := make(map[string]string, len(values))
	for key, value := range values {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "password") || strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "cvv") {
			result[key] = "[REDACTED]"
		} else {
			result[key] = value
		}
	}
	return result
}
