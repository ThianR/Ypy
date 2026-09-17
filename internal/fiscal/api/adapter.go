package api

import (
	"encoding/json"
	"errors"
	"math/big"
	"sync"
	"time"
)

type Status string

const (
	Pending  Status = "PENDING"
	Accepted Status = "ACCEPTED"
	Rejected Status = "REJECTED"
)

var ErrInvalidDocument = errors.New("invalid fiscal document")
var ErrFinalStatus = errors.New("fiscal document already finalized")
var ErrDocumentConflict = errors.New("fiscal document id reused with different content")
var ErrInvalidStatus = errors.New("fiscal status lacks required evidence")
var ErrDocumentNotAccepted = errors.New("fiscal document is not accepted")

type Document struct {
	ID, CompanyID, Payload string
	Status                 Status
	AuthorityID, Error     string
	UpdatedAt              time.Time
}

func (d Document) CanDeliver() error {
	if d.Status != Accepted || d.AuthorityID == "" {
		return ErrDocumentNotAccepted
	}
	return nil
}

type Adapter interface {
	Submit(Document) (Document, error)
	UpdateStatus(string, Status, string, string, time.Time) (Document, error)
}
type MemoryAdapter struct {
	mu        sync.Mutex
	documents map[string]Document
}

func NewMemoryAdapter() *MemoryAdapter { return &MemoryAdapter{documents: map[string]Document{}} }
func (a *MemoryAdapter) Submit(document Document) (Document, error) {
	if a == nil || document.ID == "" || document.CompanyID == "" || document.Payload == "" || !positiveInteger(document.CompanyID) || !jsonObject(document.Payload) {
		return Document{}, ErrInvalidDocument
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if previous, ok := a.documents[document.ID]; ok {
		if previous.CompanyID != document.CompanyID || previous.Payload != document.Payload {
			return Document{}, ErrDocumentConflict
		}
		return previous, nil
	}
	document.Status, document.UpdatedAt = Pending, time.Now().UTC()
	a.documents[document.ID] = document
	return document, nil
}

func positiveInteger(value string) bool {
	n, ok := new(big.Int).SetString(value, 10)
	return ok && n.Sign() > 0
}
func jsonObject(value string) bool {
	var object map[string]json.RawMessage
	return json.Unmarshal([]byte(value), &object) == nil && object != nil
}
func (a *MemoryAdapter) UpdateStatus(id string, status Status, authorityID, message string, at time.Time) (Document, error) {
	if a == nil || id == "" || (status != Accepted && status != Rejected) || at.IsZero() {
		return Document{}, ErrInvalidDocument
	}
	if (status == Accepted && authorityID == "") || (status == Rejected && message == "") {
		return Document{}, ErrInvalidStatus
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	document, ok := a.documents[id]
	if !ok {
		return Document{}, ErrInvalidDocument
	}
	if document.Status != Pending {
		return document, ErrFinalStatus
	}
	document.Status, document.AuthorityID, document.Error, document.UpdatedAt = status, authorityID, message, at.UTC()
	a.documents[id] = document
	return document, nil
}
