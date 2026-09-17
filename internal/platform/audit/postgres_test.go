package audit

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"regexp"
	"testing"
)

func TestPostgresWriterPersistsRedactedAudit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	entry, err := New("00000000-0000-0000-0000-000000000003", "ANULAR_VENTA", "motivo", map[string]string{"token": "secret"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO erp_v4.gs_auditoria")).WithArgs(int64(1), int64(2), "gv_comprobante", int64(3), "ANULAR_VENTA", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "motivo").WillReturnResult(sqlmock.NewResult(1, 1))
	if err := (PostgresWriter{DB: gormDB}).Record(context.Background(), 1, 2, 3, "gv_comprobante", entry); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
