package poscentral

import (
	"errors"
	"reflect"
	"strings"
)

var ErrSequenceGap = errors.New("SYNC_SEQUENCE_GAP")
var ErrSequenceConflict = errors.New("SYNC_SEQUENCE_CONFLICT")

type Sequencer struct {
	next    map[string]int64
	pending map[string]map[int64]Operation
	seen    map[string]map[int64]Operation
}

func NewSequencer() *Sequencer {
	return &Sequencer{next: map[string]int64{}, pending: map[string]map[int64]Operation{}, seen: map[string]map[int64]Operation{}}
}

// Push acepta operaciones solo en orden; las posteriores al hueco quedan pendientes.
func (s *Sequencer) Push(terminal string, sequence int64, operation Operation) ([]Operation, error) {
	if s == nil || strings.TrimSpace(terminal) == "" || sequence <= 0 {
		return nil, ErrSequenceConflict
	}
	if s.next[terminal] == 0 {
		s.next[terminal] = 1
	}
	if s.pending[terminal] == nil {
		s.pending[terminal] = map[int64]Operation{}
	}
	if s.seen[terminal] == nil {
		s.seen[terminal] = map[int64]Operation{}
	}
	if sequence < s.next[terminal] {
		if previous, exists := s.seen[terminal][sequence]; exists && !reflect.DeepEqual(previous, operation) {
			return nil, ErrSequenceConflict
		}
		return nil, nil
	}
	if previous, exists := s.pending[terminal][sequence]; exists && !reflect.DeepEqual(previous, operation) {
		return nil, ErrSequenceConflict
	}
	s.pending[terminal][sequence] = operation
	if sequence != s.next[terminal] {
		return nil, ErrSequenceGap
	}
	return s.drain(terminal), nil
}

func (s *Sequencer) drain(terminal string) []Operation {
	result := []Operation{}
	for {
		op, ok := s.pending[terminal][s.next[terminal]]
		if !ok {
			break
		}
		result = append(result, op)
		s.seen[terminal][s.next[terminal]] = op
		delete(s.pending[terminal], s.next[terminal])
		s.next[terminal]++
	}
	return result
}
