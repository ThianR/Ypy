package api

import (
	"errors"
	"time"
)

var ErrSessionExpired = errors.New("session expired")
var ErrInvalidSession = errors.New("invalid session")

type Session struct {
	ID, UserID, EmpresaID, SucursalID, TerminalID string
	ExpiresAt                                     time.Time
	Revoked                                       bool
}

func (s Session) Context(now time.Time) (Context, error) {
	if s.Revoked || !now.Before(s.ExpiresAt) {
		return Context{}, ErrSessionExpired
	}
	if s.ID == "" || s.UserID == "" || s.EmpresaID == "" || s.SucursalID == "" || s.TerminalID == "" {
		return Context{}, ErrInvalidSession
	}
	return Context{UserID: s.UserID, EmpresaID: s.EmpresaID, SucursalID: s.SucursalID, TerminalID: s.TerminalID, Active: true}, nil
}
