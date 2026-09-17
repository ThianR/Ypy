package poscentral

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
)

type Simulation struct {
	Seed                      int64
	Terminals, Operations     int
	DropEvery, DuplicateEvery int
}
type SimulationResult struct {
	Delivered, Gaps, Dropped, Duplicates, Recovered int
	Order                                           []int
	Trace                                           []SimulationEvent
}

type SimulationEvent struct {
	Ordinal, Sequence int
	Terminal, State   string
}

func (r SimulationResult) TraceJSON() ([]byte, error) { return json.Marshal(r.Trace) }

// Reconciled verifica que la traza final contenga exactamente una entrega
// aceptada por cada operación. Los registros duplicados y descartados son
// observaciones del transporte y no cuentan como efectos adicionales.
func (r SimulationResult) Reconciled(operations int) error {
	if operations <= 0 || r.Delivered != operations {
		return errors.New("simulation not reconciled")
	}
	seen := make(map[int]bool, operations)
	for _, event := range r.Trace {
		if event.State != "DELIVERED" && event.State != "RECOVERED" {
			continue
		}
		if event.Ordinal <= 0 || event.Ordinal > operations || seen[event.Ordinal] {
			return errors.New("duplicate simulation effect")
		}
		seen[event.Ordinal] = true
	}
	if len(seen) != operations {
		return errors.New("simulation has unresolved operations")
	}
	return nil
}

func Simulate(config Simulation) SimulationResult {
	r := rand.New(rand.NewSource(config.Seed))
	result := SimulationResult{}
	if config.Operations <= 0 {
		return result
	}
	terminals := config.Terminals
	if terminals < 1 {
		terminals = 1
	}
	type event struct {
		terminal          string
		sequence, ordinal int
	}
	events := make([]event, config.Operations)
	for i := range events {
		events[i] = event{terminal: fmt.Sprintf("terminal-%d", i%terminals), sequence: i/terminals + 1, ordinal: i + 1}
	}
	r.Shuffle(len(events), func(i, j int) { events[i], events[j] = events[j], events[i] })
	sequencer := NewSequencer()
	var dropped []event
	for _, event := range events {
		result.Order = append(result.Order, event.ordinal)
		if config.DropEvery > 0 && event.ordinal%config.DropEvery == 0 {
			result.Dropped++
			dropped = append(dropped, event)
			result.Trace = append(result.Trace, SimulationEvent{Ordinal: event.ordinal, Sequence: event.sequence, Terminal: event.terminal, State: "DROPPED"})
			continue
		}
		delivered, err := sequencer.Push(event.terminal, int64(event.sequence), Operation{ID: fmt.Sprint(event.ordinal)})
		if err == ErrSequenceGap {
			result.Gaps++
		}
		result.Delivered += len(delivered)
		result.Trace = append(result.Trace, SimulationEvent{Ordinal: event.ordinal, Sequence: event.sequence, Terminal: event.terminal, State: "DELIVERED"})
		if config.DuplicateEvery > 0 && event.ordinal%config.DuplicateEvery == 0 {
			result.Duplicates++
			result.Trace = append(result.Trace, SimulationEvent{Ordinal: event.ordinal, Sequence: event.sequence, Terminal: event.terminal, State: "DUPLICATE"})
			delivered, err = sequencer.Push(event.terminal, int64(event.sequence), Operation{ID: fmt.Sprint(event.ordinal)})
			if err == nil {
				result.Delivered += len(delivered)
			}
		}
	}
	for _, event := range dropped {
		delivered, err := sequencer.Push(event.terminal, int64(event.sequence), Operation{ID: fmt.Sprint(event.ordinal)})
		if err == nil {
			result.Recovered += len(delivered)
			result.Delivered += len(delivered)
			result.Trace = append(result.Trace, SimulationEvent{Ordinal: event.ordinal, Sequence: event.sequence, Terminal: event.terminal, State: "RECOVERED"})
		}
	}
	return result
}
