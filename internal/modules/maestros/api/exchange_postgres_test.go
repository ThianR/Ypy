package api

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"regexp"
	"testing"
	"time"
)

func newQuoteMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *gorm.DB) {
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

func TestQuoteStoreLoadsEffectiveSnapshot(t *testing.T) {
	db, mock, gormDB := newQuoteMock(t)
	defer db.Close()
	at := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT o.codigo AS from,d.codigo AS to,c.valor AS rate,c.fecha AS as_of")).WithArgs(int64(4), int64(1), int64(2), "CONTABLE", at).WillReturnRows(sqlmock.NewRows([]string{"from", "to", "rate", "as_of"}).AddRow("USD", "PYG", "7500.000000", at))
	q, err := (&QuoteStore{DB: gormDB}).LoadQuote(context.Background(), 4, 1, 2, "CONTABLE", at)
	if err != nil || q.From != "USD" || q.To != "PYG" || q.Rate != "7500.000000" {
		t.Fatalf("quote=%#v err=%v", q, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestQuoteStoreAppendsWithoutOverwritingHistory(t *testing.T) {
	db, mock, gormDB := newQuoteMock(t)
	defer db.Close()
	at := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO gs_cotizacion_moneda(empresa_id,moneda_origen_id,moneda_destino_id,fecha,tipo,valor) VALUES($1,$2,$3,$4,$5,$6)")).WithArgs(int64(4), int64(1), int64(2), at, "CONTABLE", "7500.000000").WillReturnResult(sqlmock.NewResult(1, 1))
	if err := (&QuoteStore{DB: gormDB}).SaveQuote(context.Background(), 4, 1, 2, "CONTABLE", "7500.000000", at); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestQuoteStoreRejectsNonCanonicalRate(t *testing.T) {
	if err := (&QuoteStore{}).SaveQuote(context.Background(), 4, 1, 2, "CONTABLE", "07500", time.Now()); err != ErrInvalidQuoteScope {
		t.Fatal(err)
	}
}
