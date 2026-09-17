package poscentral

import "testing"

func TestSimulationIsReproducible(t *testing.T) {
	a := Simulate(Simulation{Seed: 25, Terminals: 2, Operations: 4})
	b := Simulate(Simulation{Seed: 25, Terminals: 2, Operations: 4})
	if a.Delivered != 4 || a.Gaps == 0 {
		t.Fatal("simulation did not expose out-of-order delivery")
	}
	for i := range a.Order {
		if a.Order[i] != b.Order[i] {
			t.Fatal("same seed produced different order")
		}
	}
}

func TestSimulationRejectsEmptyWorkloadSafely(t *testing.T) {
	result := Simulate(Simulation{Operations: -1})
	if result.Delivered != 0 || result.Gaps != 0 || len(result.Order) != 0 {
		t.Fatalf("unexpected empty simulation: %#v", result)
	}
}

func TestSimulationInjectsDeterministicDropsAndDuplicates(t *testing.T) {
	a := Simulate(Simulation{Seed: 7, Terminals: 2, Operations: 6, DropEvery: 3, DuplicateEvery: 2})
	b := Simulate(Simulation{Seed: 7, Terminals: 2, Operations: 6, DropEvery: 3, DuplicateEvery: 2})
	if a.Dropped != 2 || a.Duplicates != 2 || a.Recovered == 0 || a.Delivered != 6 || a.Dropped != b.Dropped || a.Duplicates != b.Duplicates {
		t.Fatalf("unexpected deterministic fault injection: %#v %#v", a, b)
	}
}

func TestSimulationTraceIsDeterministicAndExplainsRecovery(t *testing.T) {
	a := Simulate(Simulation{Seed: 7, Terminals: 2, Operations: 6, DropEvery: 3, DuplicateEvery: 2})
	b := Simulate(Simulation{Seed: 7, Terminals: 2, Operations: 6, DropEvery: 3, DuplicateEvery: 2})
	if len(a.Trace) == 0 || len(a.Trace) != len(b.Trace) {
		t.Fatalf("missing trace: %#v", a)
	}
	seenRecovery := false
	for i, event := range a.Trace {
		if event != b.Trace[i] {
			t.Fatalf("trace differs at %d", i)
		}
		if event.State == "RECOVERED" {
			seenRecovery = true
		}
	}
	if !seenRecovery {
		t.Fatal("trace does not explain recovery")
	}
	if payload, err := a.TraceJSON(); err != nil || len(payload) == 0 {
		t.Fatalf("trace is not serializable: %v", err)
	}
	if err := a.Reconciled(6); err != nil {
		t.Fatalf("simulation did not reconcile: %v", err)
	}
}
