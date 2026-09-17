package api

import "testing"

func TestContextCannotCrossEmpresa(t *testing.T) {
	c := Context{UserID: "u1", EmpresaID: "10", SucursalID: "2", Active: true}
	if err := c.Allows("11", "2"); err != ErrUnauthorized {
		t.Fatal("cross-company access was allowed")
	}
	if err := c.Allows("10", "2"); err != nil {
		t.Fatal(err)
	}
}

func TestInactiveContextDenied(t *testing.T) {
	if err := (Context{EmpresaID: "10", SucursalID: "2"}).Allows("10", "2"); err != ErrUnauthorized {
		t.Fatal("inactive user was allowed")
	}
}
