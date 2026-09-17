package api

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"regexp"
	"testing"
)

func TestBarcodeStoreResolvesOnlyActiveSellableItems(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT b.empresa_id AS company_id,a.id AS article_id,p.id AS presentation_id,a.codigo AS article_code,a.descripcion,p.codigo AS presentation_code,p.factor_base")).WithArgs(int64(4), "779001").WillReturnRows(sqlmock.NewRows([]string{"company_id", "article_id", "presentation_id", "article_code", "description", "presentation_code", "factor_base"}).AddRow(4, 12, 20, "A-1", "Arroz", "UN", "1.000000"))
	m, err := (&BarcodeStore{DB: gormDB}).Resolve(context.Background(), 4, "779001")
	if err != nil || m.ArticleID != 12 || m.PresentationID != 20 {
		t.Fatalf("match=%#v err=%v", m, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBarcodeStoreRejectsMissingBarcode(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT").WillReturnError(sqlmock.ErrCancelled)
	if _, err := (&BarcodeStore{DB: gormDB}).Resolve(context.Background(), 0, "x"); err != ErrInvalidBarcodeScope {
		t.Fatal(err)
	}
}
