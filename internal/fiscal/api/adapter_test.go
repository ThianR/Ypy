package api

import (
	"testing"
	"time"
)

func TestFiscalPendingNeverAppearsAccepted(t *testing.T) {
	a := NewMemoryAdapter()
	doc, err := a.Submit(Document{ID: "doc-1", CompanyID: "4", Payload: "{}"})
	if err != nil || doc.Status != Pending {
		t.Fatalf("unexpected submit: %#v %v", doc, err)
	}
	if err := doc.CanDeliver(); err != ErrDocumentNotAccepted {
		t.Fatalf("pending document delivered: %v", err)
	}
	accepted, err := a.UpdateStatus("doc-1", Accepted, "AUTH-1", "", time.Now())
	if err != nil || accepted.Status != Accepted {
		t.Fatalf("unexpected accepted status: %#v %v", accepted, err)
	}
	if err := accepted.CanDeliver(); err != nil {
		t.Fatalf("accepted document blocked: %v", err)
	}
	if _, err := a.UpdateStatus("doc-1", Rejected, "", "late rejection", time.Time{}); err != ErrInvalidDocument {
		t.Fatalf("invalid timestamp accepted: %v", err)
	}
}

func TestFiscalStatusRequiresAuthorityEvidence(t *testing.T) {
	a := NewMemoryAdapter()
	if _, err := a.Submit(Document{ID: "doc-2", CompanyID: "4", Payload: "{}"}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.UpdateStatus("doc-2", Accepted, "", "", time.Now()); err != ErrInvalidStatus {
		t.Fatalf("acceptance without authority accepted: %v", err)
	}
	if _, err := a.UpdateStatus("doc-2", Rejected, "", "", time.Now()); err != ErrInvalidStatus {
		t.Fatalf("rejection without message accepted: %v", err)
	}
}
func TestFiscalRetryIsIdempotentAndFinalStatusIsImmutable(t *testing.T) {
	a := NewMemoryAdapter()
	one, _ := a.Submit(Document{ID: "doc-1", CompanyID: "4", Payload: "{}"})
	two, _ := a.Submit(Document{ID: "doc-1", CompanyID: "4", Payload: "{}"})
	if one.ID != two.ID || two.Status != Pending {
		t.Fatalf("retry changed document: %#v", two)
	}
	if _, err := a.Submit(Document{ID: "doc-1", CompanyID: "4", Payload: "{\"other\":true}"}); err != ErrDocumentConflict {
		t.Fatalf("payload conflict accepted: %v", err)
	}
	if _, err := a.UpdateStatus("doc-1", Rejected, "", "bad payload", time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := a.UpdateStatus("doc-1", Accepted, "AUTH-2", "", time.Now()); err != ErrFinalStatus {
		t.Fatalf("final status changed: %v", err)
	}
}

func TestFiscalRejectsInvalidPayloadAndScope(t *testing.T) {
	a := NewMemoryAdapter()
	for _, document := range []Document{{ID: "d1", CompanyID: "0", Payload: "{}"}, {ID: "d2", CompanyID: "4", Payload: "[]"}, {ID: "d3", CompanyID: "4", Payload: "not-json"}} {
		if _, err := a.Submit(document); err != ErrInvalidDocument {
			t.Fatalf("invalid fiscal document accepted: %#v %v", document, err)
		}
	}
}
