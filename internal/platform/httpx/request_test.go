package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDIsStable(t *testing.T) {
	wanted := "00000000-0000-0000-0000-000000000001"
	h := WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if RequestID(r.Context()) != wanted {
			t.Fatal("request id not propagated")
		}
	}))
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Request-ID", wanted)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	if rec.Header().Get("X-Request-ID") != wanted {
		t.Fatal("request id not returned")
	}
}

func TestRequestIDRejectsInvalidHeaderAndReturnsGeneratedID(t *testing.T) {
	h := WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := json.Marshal(RequestID(r.Context())); err != nil || RequestID(r.Context()) == "" {
			t.Fatal("expected generated request id")
		}
	}))
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Request-ID", "not-a-uuid")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	if rec.Header().Get("X-Request-ID") == "not-a-uuid" || rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("invalid request id was returned")
	}
}

func TestWriteErrorUsesStableJSONContract(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteError(rec, http.StatusConflict, "CONFLICT", "conflict", "request-1", true)
	if rec.Code != http.StatusConflict || rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected response metadata: %d %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	var fields map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &fields); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"code", "message", "correlation_id", "retryable"} {
		if _, ok := fields[key]; !ok {
			t.Fatalf("missing JSON field %q: %s", key, rec.Body.String())
		}
	}
	var got Error
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Code != "CONFLICT" || got.Message != "conflict" || got.CorrelationID != "request-1" || !got.Retryable {
		t.Fatalf("unexpected error body: %+v", got)
	}
}
