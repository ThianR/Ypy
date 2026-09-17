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

func newVentasMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *gorm.DB) {
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

func TestReturnStoreDelegatesToV4(t *testing.T) {
	db, mock, gormDB := newVentasMock(t)
	defer db.Close()
	id := uuid.MustParse("00000000-0000-0000-0000-000000000016")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT erp_v4.gv_devolver($1,$2,$3,$4,$5)")).WithArgs(int64(4), int64(70), "2.000000", id, "cliente solicita cambio").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(92))
	got, err := (&ReturnStore{DB: gormDB}).Create(context.Background(), 4, 70, "2.000000", " cliente solicita cambio ", id)
	if err != nil || got != 92 {
		t.Fatalf("return=%d err=%v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreditNoteStorePersistsAndConfirms(t *testing.T) {
	db, mock, gormDB := newVentasMock(t)
	defer db.Close()
	id := uuid.MustParse("00000000-0000-0000-0000-000000000017")
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO gv_nota_credito_cabecera(empresa_id,uid,sucursal_id,usuario_id,tipo_id,tercero_id,moneda_id,fecha_operacion,total,motivo_excepcion,comprobante_origen_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id")).WithArgs(int64(4), id, int64(2), int64(7), int64(8), int64(9), int64(1), "2026-09-16T12:00:00Z", "100.000000", "devolucion", int64(50)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(93))
	got, err := (&CreditNoteStore{DB: gormDB}).Create(context.Background(), CreditNote{ID: id, CompanyID: 4, BranchID: 2, UserID: 7, TypeID: 8, CustomerID: 9, CurrencyID: 1, OriginID: 50, Total: "100.000000", Date: "2026-09-16T12:00:00Z", Reason: " devolucion "})
	if err != nil || got != 93 {
		t.Fatalf("note=%d err=%v", got, err)
	}
	mock.ExpectExec(regexp.QuoteMeta("SELECT gs_confirmar($1,$2,$3)")).WithArgs(int64(4), "gv_nota_credito", int64(93)).WillReturnResult(sqlmock.NewResult(0, 0))
	if err := (&CreditNoteStore{DB: gormDB}).Confirm(context.Background(), 4, 93); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
