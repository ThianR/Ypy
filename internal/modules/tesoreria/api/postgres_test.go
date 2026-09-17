package api

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newTreasuryMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *gorm.DB) {
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

func TestPostgresStoreAddsCashMovementIdempotently(t *testing.T) {
	db, mock, gormDB := newTreasuryMock(t)
	defer db.Close()
	id := uuid.MustParse("00000000-0000-0000-0000-000000000015")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,empresa_id,apertura_id,medio_id,tipo,importe::text,motivo,fecha::text FROM erp_v4.gf_movimiento_caja WHERE uid=$1")).WithArgs(id).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO erp_v4.gf_movimiento_caja(empresa_id,uid,apertura_id,medio_id,tipo,importe,motivo,fecha) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id")).WithArgs(int64(4), id, int64(33), int64(5), "RETIRO", "10.000000", "cambio", "2026-09-16T12:00:00Z").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(91))
	got, err := (&PostgresStore{DB: gormDB}).AddMovement(context.Background(), CashMovement{ID: id, CompanyID: 4, OpeningID: 33, MethodID: 5, Kind: "RETIRO", Amount: "10.000000", Reason: " cambio ", Date: "2026-09-16T12:00:00Z"})
	if err != nil || got != 91 {
		t.Fatalf("movement=%d err=%v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresCashOpenDelegatesToV4(t *testing.T) {
	db, mock, gormDB := newTreasuryMock(t)
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO erp_v4.gf_caja_apertura(empresa_id,caja_id,usuario_id,terminal_id,fondo) VALUES($1,$2,$3,$4,$5) RETURNING id")).WithArgs(int64(4), int64(2), int64(7), int64(9), "100.000000").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(33))
	id, err := (&PostgresStore{DB: gormDB}).Open(context.Background(), 4, 2, 7, 9, "100.000000")
	if err != nil || id != 33 {
		t.Fatalf("id=%d err=%v", id, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresCashCloseDelegatesCountsToV4(t *testing.T) {
	db, mock, gormDB := newTreasuryMock(t)
	defer db.Close()
	mock.ExpectExec(regexp.QuoteMeta("SELECT erp_v4.gf_cerrar_caja($1,$2,$3::jsonb)")).WithArgs(int64(4), int64(33), `{"10":"100.000000"}`).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := (&PostgresStore{DB: gormDB}).Close(context.Background(), 4, 33, []byte(`{"10":"100.000000"}`)); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	if err := (&PostgresStore{DB: gormDB}).Close(context.Background(), 4, 33, []byte("not-json")); err != ErrInvalidCashScope {
		t.Fatalf("expected invalid counts, got %v", err)
	}
}
