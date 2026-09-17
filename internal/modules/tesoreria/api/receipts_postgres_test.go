package api

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"regexp"
	"testing"
)

func TestReceiptStorePersistsReceiptAndForm(t *testing.T) {
	db, mock, gormDB := newTreasuryMock(t)
	defer db.Close()
	id := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,empresa_id,tercero_id,apertura_id,moneda_id,importe::text,fecha::text FROM gf_recibo_cabecera WHERE uid=$1")).WithArgs(id).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO gf_recibo_cabecera(empresa_id,uid,tercero_id,apertura_id,moneda_id,importe,fecha) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id")).WithArgs(int64(4), id, int64(20), int64(33), int64(1), "100.000000", "2026-09-16T12:00:00Z").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(80))
	h, err := (&ReceiptStore{DB: gormDB}).Create(context.Background(), Receipt{ID: id, CompanyID: 4, CustomerID: 20, OpeningID: 33, CurrencyID: 1, Amount: "100.000000", Date: "2026-09-16T12:00:00Z"})
	if err != nil || h != 80 {
		t.Fatalf("header=%d err=%v", h, err)
	}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO gf_recibo_forma(empresa_id,cabecera_id,medio_id,importe,referencia) VALUES($1,$2,$3,$4,$5)")).WithArgs(int64(4), int64(80), int64(5), "100.000000", "ref").WillReturnResult(sqlmock.NewResult(1, 1))
	if err := (&ReceiptStore{DB: gormDB}).AddForm(context.Background(), ReceiptForm{CompanyID: 4, HeaderID: 80, MethodID: 5, Amount: "100.000000", Reference: "ref"}); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReceiptStoreRejectsUUIDContentConflict(t *testing.T) {
	db, mock, gormDB := newTreasuryMock(t)
	defer db.Close()
	id := uuid.MustParse("00000000-0000-0000-0000-000000000011")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,empresa_id,tercero_id,apertura_id,moneda_id,importe::text,fecha::text FROM gf_recibo_cabecera WHERE uid=$1")).WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"id", "empresa", "customer", "opening", "currency", "amount", "date"}).AddRow(80, 4, 20, 33, 1, "100.000000", "2026-09-16T12:00:00Z"))
	if _, err := (&ReceiptStore{DB: gormDB}).Create(context.Background(), Receipt{ID: id, CompanyID: 4, CustomerID: 20, OpeningID: 33, CurrencyID: 1, Amount: "101.000000", Date: "2026-09-16T12:00:00Z"}); err != ErrReceiptConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReceiptStoreConfirmsThroughV4(t *testing.T) {
	db, mock, gormDB := newTreasuryMock(t)
	defer db.Close()
	mock.ExpectExec(regexp.QuoteMeta("SELECT gf_confirmar($1,$2,$3)")).WithArgs(int64(4), "gf_recibo", int64(80)).WillReturnResult(sqlmock.NewResult(0, 0))
	if err := (&ReceiptStore{DB: gormDB}).Confirm(context.Background(), 4, 80); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReceiptStoreCreatesMixedReceiptAtomically(t *testing.T) {
	db, mock, gormDB := newTreasuryMock(t)
	defer db.Close()
	id := uuid.MustParse("00000000-0000-0000-0000-000000000012")
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,empresa_id,tercero_id,apertura_id,moneda_id,importe::text,fecha::text,estado FROM gf_recibo_cabecera WHERE uid=$1 FOR UPDATE")).WithArgs(id).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO gf_recibo_cabecera(empresa_id,uid,tercero_id,apertura_id,moneda_id,importe,fecha) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id")).WithArgs(int64(4), id, int64(20), int64(33), int64(1), "100.000000", "2026-09-16T12:00:00Z").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(81))
	formSQL := regexp.QuoteMeta("INSERT INTO gf_recibo_forma(empresa_id,cabecera_id,medio_id,importe,referencia) VALUES($1,$2,$3,$4,$5)")
	mock.ExpectExec(formSQL).WithArgs(int64(4), int64(81), int64(5), "60.000000", "efectivo").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(formSQL).WithArgs(int64(4), int64(81), int64(6), "40.000000", "tarjeta").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COALESCE(SUM(importe),0) = $1::numeric FROM gf_recibo_forma WHERE empresa_id=$2 AND cabecera_id=$3")).WithArgs("100.000000", int64(4), int64(81)).WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(true))
	mock.ExpectExec(regexp.QuoteMeta("SELECT gf_confirmar($1,$2,$3)")).WithArgs(int64(4), "gf_recibo", int64(81)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
	got, err := (&ReceiptStore{DB: gormDB}).CreateAndConfirm(context.Background(), Receipt{ID: id, CompanyID: 4, CustomerID: 20, OpeningID: 33, CurrencyID: 1, Amount: "100.000000", Date: "2026-09-16T12:00:00Z"}, []ReceiptForm{{CompanyID: 4, MethodID: 5, Amount: "60.000000", Reference: "efectivo"}, {CompanyID: 4, MethodID: 6, Amount: "40.000000", Reference: "tarjeta"}})
	if err != nil || got != 81 {
		t.Fatalf("header=%d err=%v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReceiptStoreReturnsConfirmedReceiptOnRetry(t *testing.T) {
	db, mock, gormDB := newTreasuryMock(t)
	defer db.Close()
	id := uuid.MustParse("00000000-0000-0000-0000-000000000013")
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,empresa_id,tercero_id,apertura_id,moneda_id,importe::text,fecha::text,estado FROM gf_recibo_cabecera WHERE uid=$1 FOR UPDATE")).WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"id", "empresa", "customer", "opening", "currency", "amount", "date", "state"}).AddRow(82, 4, 20, 33, 1, "100.000000", "2026-09-16T12:00:00Z", "CONFIRMADO"))
	mock.ExpectCommit()
	got, err := (&ReceiptStore{DB: gormDB}).CreateAndConfirm(context.Background(), Receipt{ID: id, CompanyID: 4, CustomerID: 20, OpeningID: 33, CurrencyID: 1, Amount: "100.000000", Date: "2026-09-16T12:00:00Z"}, []ReceiptForm{{CompanyID: 4, MethodID: 5, Amount: "100.000000"}})
	if err != nil || got != 82 {
		t.Fatalf("header=%d err=%v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
