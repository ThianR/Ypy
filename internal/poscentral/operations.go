package poscentral

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

var ErrConflict = errors.New("operation content conflict")
var ErrInvalidOperation = errors.New("invalid operation")

type Operation struct {
	ID         string `json:"id"`
	EmpresaID  string `json:"empresa_id"`
	TerminalID string `json:"terminal_id"`
	Payload    string `json:"payload"`
}
type Store struct {
	mu         sync.Mutex
	operations map[string]Operation
}

func NewStore() *Store { return &Store{operations: map[string]Operation{}} }

func (s *Store) Apply(op Operation) (Operation, error) {
	if s == nil {
		return Operation{}, ErrInvalidOperation
	}
	if err := Validate(op); err != nil {
		return Operation{}, err
	}
	op.Payload = NormalizePayload(op.Payload)
	s.mu.Lock()
	defer s.mu.Unlock()
	if previous, ok := s.operations[op.ID]; ok {
		if previous.EmpresaID != op.EmpresaID || previous.TerminalID != op.TerminalID || hash(previous.Payload) != hash(op.Payload) {
			return Operation{}, ErrConflict
		}
		return previous, nil
	}
	s.operations[op.ID] = op
	return op, nil
}

func Validate(op Operation) error {
	payload := strings.TrimSpace(op.Payload)
	company, companyErr := strconv.ParseInt(op.EmpresaID, 10, 64)
	terminal, terminalErr := strconv.ParseInt(op.TerminalID, 10, 64)
	if _, err := uuid.Parse(op.ID); err != nil || companyErr != nil || company <= 0 || terminalErr != nil || terminal <= 0 || payload == "" || !json.Valid([]byte(payload)) || payload[0] != '{' {
		return ErrInvalidOperation
	}
	return nil
}

func hash(value string) string {
	sum := sha256.Sum256([]byte(NormalizePayload(value)))
	return hex.EncodeToString(sum[:])
}

func ContentHash(value string) string { return hash(value) }

// NormalizePayload compacts JSON so equivalent representations share one hash.
func NormalizePayload(value string) string {
	var decoded any
	if err := json.Unmarshal([]byte(value), &decoded); err != nil {
		return value
	}
	var compacted bytes.Buffer
	if err := json.NewEncoder(&compacted).Encode(decoded); err != nil {
		return value
	}
	return strings.TrimSpace(compacted.String())
}

func Decode(data []byte) (Operation, error) {
	var op Operation
	err := json.Unmarshal(data, &op)
	if err == nil {
		err = Validate(op)
		if err == nil {
			op.Payload = NormalizePayload(op.Payload)
		}
	}
	return op, err
}
