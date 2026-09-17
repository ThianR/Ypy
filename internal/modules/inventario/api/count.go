package api

import (
	"errors"
	"strings"
	"sync"
)

var ErrAdjustmentUnauthorized = errors.New("adjustment authorization required")
var ErrAdjustmentAlreadyApplied = errors.New("adjustment already applied")
var ErrInvalidCount = errors.New("invalid physical count")

type PhysicalCount struct {
	ID, ItemID, Reason string
	Reference, Counted int64
	Applied            bool
	mu                 sync.Mutex
}

func (c *PhysicalCount) Difference() int64 {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Counted - c.Reference
}
func (c *PhysicalCount) Approve(supervisor, reason string) error {
	if c == nil {
		return ErrInvalidCount
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Applied {
		return ErrAdjustmentAlreadyApplied
	}
	if strings.TrimSpace(supervisor) == "" || strings.TrimSpace(reason) == "" {
		return ErrAdjustmentUnauthorized
	}
	if c.ID == "" || c.ItemID == "" || c.Reference < 0 || c.Counted < 0 {
		return ErrInvalidCount
	}
	c.Reason = strings.TrimSpace(reason)
	c.Applied = true
	return nil
}
