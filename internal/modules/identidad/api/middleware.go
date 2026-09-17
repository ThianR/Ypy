package api

import (
	"context"
	"net/http"
	"time"
)

type sessionContextKey struct{}

type ContextLoader interface {
	LoadContext(context.Context, string, time.Time) (Context, error)
}

func RequestContext(ctx context.Context) (Context, bool) {
	value, ok := ctx.Value(sessionContextKey{}).(Context)
	return value, ok
}

// WithSession requiere una sesión persistente y coloca su contexto de ámbito
// en la solicitud para las comprobaciones posteriores de autorización.
func WithSession(next http.Handler, loader ContextLoader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if loader == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		sessionID := r.Header.Get("X-Session-ID")
		ctx, err := loader.LoadContext(r.Context(), sessionID, time.Now().UTC())
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), sessionContextKey{}, ctx)))
	})
}
