package api

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"regexp"
	"testing"
)

func TestPostgresReceiptConfirmDelegatesToV4(t *testing.T) {
	db, mock, gormDB := newComprasMock(t)
	defer db.Close()
	mock.ExpectExec(regexp.QuoteMeta("SELECT gs_confirmar($1,$2,$3)")).WithArgs(int64(4), "gc_recepcion", int64(55)).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := (&PostgresStore{DB: gormDB}).Confirm(context.Background(), 4, 55); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
