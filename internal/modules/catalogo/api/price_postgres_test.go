package api

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"regexp"
	"testing"
	"time"
)

func TestBarcodeStoreLoadsPriceWithinEffectiveWindow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT precio::text FROM erp_v4.gv_precio")).WithArgs(int64(4), int64(2), int64(20), at, "2").WillReturnRows(sqlmock.NewRows([]string{"precio"}).AddRow("6500.000000"))
	price, err := (&BarcodeStore{DB: gormDB}).Price(context.Background(), 4, 2, 20, "2", at)
	if err != nil || price != "6500.000000" {
		t.Fatalf("price=%q err=%v", price, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
