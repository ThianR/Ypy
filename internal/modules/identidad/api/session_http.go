package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type SessionManager interface {
	LoadContext(context.Context, string, time.Time) (Context, error)
	Revoke(context.Context, string, time.Time) error
}

// SessionHandler permite verificar y cerrar la sesión desde la interfaz web.
func SessionHandler(store SessionManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodDelete {
			w.Header().Set("Allow", "GET, DELETE")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if store == nil {
			http.Error(w, "El servicio de acceso no está disponible.", http.StatusServiceUnavailable)
			return
		}
		id := r.Header.Get("X-Session-ID")
		scope, err := store.LoadContext(r.Context(), id, time.Now().UTC())
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		if r.Method == http.MethodDelete {
			if err := store.Revoke(r.Context(), id, time.Now().UTC()); err != nil {
				http.Error(w, "No se pudo cerrar la sesión.", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"usuario_id": scope.UserID, "empresa_id": scope.EmpresaID,
			"sucursal_id": scope.SucursalID, "terminal_id": scope.TerminalID,
		})
	})
}
