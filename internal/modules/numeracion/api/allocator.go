package api

import (
	"errors"
	"time"
)

var ErrRangeExhausted = errors.New("number range exhausted")
var ErrWrongTerminal = errors.New("number range belongs to another terminal")
var ErrInvalidRange = errors.New("invalid number range")
var ErrPermissionExpired = errors.New("number permission expired")

type Block struct {
	TerminalID string
	Next, End  int64
	ExpiresAt  time.Time
}

func (b *Block) Consume(terminalID string) (int64, error) {
	return b.ConsumeAt(terminalID, time.Now())
}

func (b *Block) ConsumeAt(terminalID string, at time.Time) (int64, error) {
	if b == nil || b.TerminalID == "" || terminalID == "" || b.Next <= 0 || b.End < b.Next {
		if b != nil && b.End < b.Next {
			return 0, ErrRangeExhausted
		}
		return 0, ErrInvalidRange
	}
	if b.TerminalID != terminalID {
		return 0, ErrWrongTerminal
	}
	if !b.ExpiresAt.IsZero() && !at.Before(b.ExpiresAt) {
		return 0, ErrPermissionExpired
	}
	if b.Next > b.End {
		return 0, ErrRangeExhausted
	}
	n := b.Next
	b.Next++
	return n, nil
}

type Allocator struct {
	next   int64
	blocks map[string][]*Block
}

func NewAllocator(first int64) *Allocator {
	return &Allocator{next: first, blocks: map[string][]*Block{}}
}

func (a *Allocator) Reserve(terminalID string, size int64) *Block {
	return a.ReserveUntil(terminalID, size, time.Time{})
}

// ReserveUntil crea un bloque exclusivo y opcionalmente limitado por vencimiento.
func (a *Allocator) ReserveUntil(terminalID string, size int64, expiresAt time.Time) *Block {
	if a == nil || terminalID == "" || size <= 0 || a.next <= 0 || size > (int64(^uint64(0)>>1)-a.next+1) {
		return nil
	}
	b := &Block{TerminalID: terminalID, Next: a.next, End: a.next + size - 1, ExpiresAt: expiresAt}
	a.next += size
	a.blocks[terminalID] = append(a.blocks[terminalID], b)
	return b
}
