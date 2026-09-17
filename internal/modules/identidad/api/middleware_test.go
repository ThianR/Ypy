package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type testLoader struct{ err error }

func (l testLoader) LoadContext(context.Context, string, time.Time) (Context, error) {
	if l.err != nil {
		return Context{}, l.err
	}
	return Context{UserID: "7", EmpresaID: "4", SucursalID: "2", TerminalID: "9", Active: true}, nil
}

func TestWithSessionInjectsContext(t *testing.T) {
	handler := WithSession(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, ok := RequestContext(r.Context())
		if !ok || ctx.EmpresaID != "4" {
			t.Fatal("authorized context was not injected")
		}
		w.WriteHeader(http.StatusNoContent)
	}), testLoader{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Session-ID", "session")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestWithSessionRejectsMissingOrInvalidSession(t *testing.T) {
	handler := WithSession(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), testLoader{err: ErrSessionExpired})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rec.Code)
	}
}
