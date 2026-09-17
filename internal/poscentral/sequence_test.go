package poscentral

import "testing"

func TestSequencerRejectsInvalidScopeAndSequence(t *testing.T) {
	s := NewSequencer()
	for _, terminal := range []string{"", "  "} {
		if _, err := s.Push(terminal, 1, Operation{}); err != ErrSequenceConflict {
			t.Fatalf("expected invalid terminal rejection, got %v", err)
		}
	}
	if _, err := s.Push("terminal-1", 0, Operation{}); err != ErrSequenceConflict {
		t.Fatalf("expected invalid sequence rejection, got %v", err)
	}
	var nilSequencer *Sequencer
	if _, err := nilSequencer.Push("terminal-1", 1, Operation{}); err != ErrSequenceConflict {
		t.Fatalf("expected nil sequencer rejection, got %v", err)
	}
}

func TestSequencerRetainsGap(t *testing.T) {
	s := NewSequencer()
	if _, err := s.Push("t1", 2, Operation{ID: "b"}); err != ErrSequenceGap {
		t.Fatal("gap was not reported")
	}
	got, err := s.Push("t1", 1, Operation{ID: "a"})
	if err != nil || len(got) != 2 || got[0].ID != "a" || got[1].ID != "b" {
		t.Fatalf("unexpected drain: %#v %v", got, err)
	}
}

func TestSequenceConflictDoesNotOverwritePendingOperation(t *testing.T) {
	s := NewSequencer()
	if _, err := s.Push("t1", 2, Operation{ID: "first"}); err != ErrSequenceGap {
		t.Fatalf("expected gap, got %v", err)
	}
	if _, err := s.Push("t1", 2, Operation{ID: "second"}); err != ErrSequenceConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
	got, err := s.Push("t1", 1, Operation{ID: "one"})
	if err != nil || len(got) != 2 || got[1].ID != "first" {
		t.Fatalf("pending operation was overwritten: %#v %v", got, err)
	}
}

func TestSequenceConflictAfterDelivery(t *testing.T) {
	s := NewSequencer()
	if got, err := s.Push("t1", 1, Operation{ID: "first"}); err != nil || len(got) != 1 {
		t.Fatalf("unexpected first delivery: %#v %v", got, err)
	}
	if _, err := s.Push("t1", 1, Operation{ID: "changed"}); err != ErrSequenceConflict {
		t.Fatalf("expected delivered conflict, got %v", err)
	}
}
