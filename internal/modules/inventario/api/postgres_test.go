package api

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestBalanceCurrentReadsAuthoritativeV4Positions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COALESCE(SUM(e.cantidad),0)::bigint")).WithArgs(int64(4), int64(12)).WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(int64(18)))
	got, err := (&PostgresStore{DB: gormDB}).BalanceCurrent(context.Background(), 4, 12)
	if err != nil || got != 18 {
		t.Fatalf("balance=%d err=%v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBalanceCurrentRejectsUnscopedIDs(t *testing.T) {
	if _, err := (&PostgresStore{}).BalanceCurrent(context.Background(), 0, 12); err != ErrInvalidScope {
		t.Fatal(err)
	}
}
