package api

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"regexp"
	"testing"
	"time"
)

func newComprasMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *gorm.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return db, mock, gormDB
}

func TestPostgresReceiptHeaderIsIdempotentByUUID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse("00000000-0000-0000-0000-000000000009")
	at := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,empresa_id,sucursal_id,usuario_id,tipo_id,terminal_id,tercero_id,moneda_id,fecha_operacion,total,tercero_snapshot::text FROM gc_recepcion_cabecera WHERE uid=$1")).WithArgs(id).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO gc_recepcion_cabecera(empresa_id,uid,sucursal_id,usuario_id,tipo_id,terminal_id,tercero_id,moneda_id,fecha_operacion,total,tercero_snapshot) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id")).WithArgs(int64(4), id, int64(1), int64(7), int64(2), int64(9), int64(20), int64(1), at, "100.000000", "{}").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(55))
	got, err := (&PostgresStore{DB: gormDB}).CreateHeader(context.Background(), ReceiptHeader{ID: id, CompanyID: 4, BranchID: 1, UserID: 7, TypeID: 2, TerminalID: 9, SupplierID: 20, CurrencyID: 1, Total: "100.000000", Snapshot: "{}", At: at})
	if err != nil || got != 55 {
		t.Fatalf("id=%d err=%v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
