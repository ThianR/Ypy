package api

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"regexp"
	"testing"
)

func TestPostgresReceiptDetailIsIdempotentByLine(t *testing.T) {
	db, mock, gormDB := newComprasMock(t)
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,articulo_id,presentacion_id,descripcion,cantidad::text,factor_base::text,precio::text FROM gc_recepcion_detalle")).WithArgs(int64(4), int64(55), int64(1)).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO gc_recepcion_detalle(empresa_id,cabecera_id,renglon,articulo_id,presentacion_id,descripcion,cantidad,factor_base,precio) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id")).WithArgs(int64(4), int64(55), int64(1), int64(12), int64(20), "Arroz", "10", "1.000000", "6500.000000").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(70))
	id, err := (&PostgresStore{DB: gormDB}).AddDetail(context.Background(), ReceiptDetail{CompanyID: 4, HeaderID: 55, Line: 1, ArticleID: 12, PresentationID: 20, Description: "Arroz", Quantity: "10", FactorBase: "1.000000", Price: "6500.000000"})
	if err != nil || id != 70 {
		t.Fatalf("id=%d err=%v", id, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
