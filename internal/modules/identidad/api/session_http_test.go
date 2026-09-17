package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type browserSession struct{ revoked bool }

func (s *browserSession) LoadContext(_ context.Context, id string, _ time.Time) (Context, error) {
	if id != "sesion-valida" || s.revoked {
		return Context{}, ErrSessionExpired
	}
	return Context{UserID: "1", EmpresaID: "1", SucursalID: "1", TerminalID: "1", Active: true}, nil
}

func (s *browserSession) Revoke(_ context.Context, _ string, _ time.Time) error {
	s.revoked = true
	return nil
}

func TestBrowserSessionLifecycle(t *testing.T) {
	store := &browserSession{}
	handler := SessionHandler(store)
	for _, test := range []struct {
		method, id string
		status     int
	}{
		{http.MethodGet, "", http.StatusUnauthorized},
		{http.MethodGet, "sesion-valida", http.StatusOK},
		{http.MethodDelete, "sesion-valida", http.StatusNoContent},
		{http.MethodGet, "sesion-valida", http.StatusUnauthorized},
	} {
		request := httptest.NewRequest(test.method, "/api/v1/auth/session", nil)
		request.Header.Set("X-Session-ID", test.id)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != test.status {
			t.Fatalf("Estado esperado %d, recibido %d", test.status, response.Code)
		}
	}
}

func TestBrowserSessionUnavailable(t *testing.T) {
	response := httptest.NewRecorder()
	SessionHandler(nil).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatal("Debe informar que la autenticación no está disponible")
	}
}
