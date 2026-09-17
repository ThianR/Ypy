package outbox

import (
	"encoding/json"
	"errors"
)

var ErrTemporary = errors.New("temporary publish failure")
var ErrInvalidWorker = errors.New("invalid outbox worker input")

type Event struct {
	ID, Topic, Payload string
	Published          bool
	Retries            int
}
type Publisher interface{ Publish(Event) error }

// Process publica cada evento pendiente; los fallos temporales quedan pendientes para reintento.
func Process(events []*Event, publisher Publisher) error {
	if publisher == nil {
		return ErrInvalidWorker
	}
	for _, event := range events {
		if event == nil || event.ID == "" || event.Topic == "" || !json.Valid([]byte(event.Payload)) {
			return ErrInvalidWorker
		}
		if event.Published {
			continue
		}
		if err := publisher.Publish(*event); err != nil {
			event.Retries++
			return err
		}
		event.Published = true
	}
	return nil
}
