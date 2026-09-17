package api

import "errors"

var ErrUnauthorized = errors.New("unauthorized context")

type Context struct {
	UserID, EmpresaID, SucursalID, TerminalID string
	Active                                    bool
	Permissions                               map[string]bool
}

func (c Context) Can(permission, empresaID, sucursalID string) error {
	if err := c.Allows(empresaID, sucursalID); err != nil {
		return err
	}
	if !c.Permissions[permission] {
		return ErrUnauthorized
	}
	return nil
}

func (c Context) Allows(empresaID, sucursalID string) error {
	if !c.Active || c.EmpresaID != empresaID || c.SucursalID != sucursalID {
		return ErrUnauthorized
	}
	return nil
}
