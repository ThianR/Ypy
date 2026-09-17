package api

import "testing"

func TestPermissionRequiresScopeAndGrant(t *testing.T) {
	c := Context{EmpresaID: "1", SucursalID: "2", Active: true, Permissions: map[string]bool{"sale.confirm": true}}
	if err := c.Can("sale.confirm", "1", "2"); err != nil {
		t.Fatal(err)
	}
	if err := c.Can("sale.confirm", "9", "2"); err != ErrUnauthorized {
		t.Fatal("cross-company permission accepted")
	}
	if err := c.Can("cash.close", "1", "2"); err != ErrUnauthorized {
		t.Fatal("ungranted permission accepted")
	}
}
