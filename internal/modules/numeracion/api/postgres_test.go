package api

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newNumberingMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *gorm.DB) {
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

func TestPostgresStoreDelegatesExclusiveRangeReservation(t *testing.T) {
	db, mock, gormDB := newNumberingMock(t)
	defer db.Close()
	key := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT gs_reservar_rango($1,$2,$3,$4,$5,$6,$7)")).WithArgs(int64(1), int64(9), int64(3), int64(7), int64(10), key, false).WillReturnRows(sqlmock.NewRows([]string{"gs_reservar_rango"}).AddRow(44))
	id, err := (&PostgresStore{DB: gormDB}).Reserve(context.Background(), 1, 9, 3, 7, 10, key, false)
	if err != nil || id != 44 {
		t.Fatalf("reserve=%d err=%v", id, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresStoreAssignsFromTerminalRange(t *testing.T) {
	db, mock, gormDB := newNumberingMock(t)
	defer db.Close()
	key := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	at := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT gs_asignar_numero($1,$2,$3,$4,$5,$6,$7,$8)")).WithArgs(int64(1), int64(9), int64(7), key, int64(3), int64(44), int64(101), at).WillReturnRows(sqlmock.NewRows([]string{"gs_asignar_numero"}).AddRow(88))
	id, err := (&PostgresStore{DB: gormDB}).Assign(context.Background(), 1, 9, 7, 3, 44, 101, key, at)
	if err != nil || id != 88 {
		t.Fatalf("assign=%d err=%v", id, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresStoreRejectsInvalidScope(t *testing.T) {
	if _, err := (&PostgresStore{}).Reserve(context.Background(), 0, 1, 1, 1, 1, uuid.New(), false); err != ErrInvalidRequest {
		t.Fatal(err)
	}
}
