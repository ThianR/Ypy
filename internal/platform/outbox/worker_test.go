package outbox

import "testing"

type publisher struct {
	calls    int
	failures int
}

func (p *publisher) Publish(Event) error {
	p.calls++
	if p.failures > 0 {
		p.failures--
		return ErrTemporary
	}
	return nil
}

func TestOutboxRetriesWithoutDuplicatingPublished(t *testing.T) {
	events := []*Event{{ID: "1", Topic: "SaleConfirmed", Payload: "{}"}, {ID: "2", Topic: "PaymentRecorded", Payload: "{}", Published: true}}
	p := &publisher{failures: 1}
	if err := Process(events, p); err != ErrTemporary || events[0].Retries != 1 {
		t.Fatal("temporary failure was not retained")
	}
	if err := Process(events, p); err != nil || !events[0].Published {
		t.Fatal(err)
	}
	if p.calls != 2 {
		t.Fatalf("published event was retried: %d calls", p.calls)
	}
}

func TestOutboxRejectsInvalidWorkerInput(t *testing.T) {
	if err := Process([]*Event{{ID: "1", Topic: "x", Payload: "bad"}}, &publisher{}); err != ErrInvalidWorker {
		t.Fatalf("invalid payload accepted: %v", err)
	}
	if err := Process(nil, nil); err != ErrInvalidWorker {
		t.Fatalf("nil publisher accepted: %v", err)
	}
}
