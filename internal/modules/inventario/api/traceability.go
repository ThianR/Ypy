package api

import (
	"errors"
	"strings"
)

var ErrDuplicateSerial = errors.New("serial already registered")
var ErrUnknownSerial = errors.New("serial not registered")
var ErrTraceabilityRequired = errors.New("traceability is required")

type TrackedItem struct {
	ItemID               string
	RequiresTraceability bool
	serials              map[string]struct{}
}

func NewTrackedItem(id string, required bool) *TrackedItem {
	return &TrackedItem{ItemID: id, RequiresTraceability: required, serials: map[string]struct{}{}}
}
func (i *TrackedItem) RegisterSerial(serial string) error {
	serial = strings.TrimSpace(serial)
	if i == nil || i.ItemID == "" || serial == "" {
		return ErrUnknownSerial
	}
	if _, ok := i.serials[serial]; ok {
		return ErrDuplicateSerial
	}
	i.serials[serial] = struct{}{}
	return nil
}
func (i *TrackedItem) ValidateSale(serials []string) error {
	if i == nil || i.ItemID == "" {
		return ErrUnknownSerial
	}
	if i.RequiresTraceability && len(serials) == 0 {
		return ErrTraceabilityRequired
	}
	seen := make(map[string]struct{}, len(serials))
	for _, serial := range serials {
		serial = strings.TrimSpace(serial)
		if serial == "" {
			return ErrUnknownSerial
		}
		if _, ok := seen[serial]; ok {
			return ErrDuplicateSerial
		}
		seen[serial] = struct{}{}
		if _, ok := i.serials[serial]; !ok {
			return ErrUnknownSerial
		}
	}
	return nil
}
