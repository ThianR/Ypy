package outbox

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newOutboxMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *gorm.DB) {
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

func TestPostgresStorePendingAndMarkPublished(t *testing.T) {
	db, mock, gormDB := newOutboxMock(t)
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, tema, contenido::text FROM erp_v4.gs_evento_salida")).
		WithArgs(10).WillReturnRows(sqlmock.NewRows([]string{"id", "tema", "contenido"}).AddRow(7, "venta.confirmada", `{"id":7}`))
	events, err := (PostgresStore{DB: gormDB}).Pending(context.Background(), 10)
	if err != nil || len(events) != 1 || events[0].ID != "7" {
		t.Fatalf("unexpected pending events: %+v %v", events, err)
	}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE erp_v4.gs_evento_salida SET publicado_en=now()")).
		WithArgs(int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := (PostgresStore{DB: gormDB}).MarkPublished(context.Background(), "7"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestValidatePayloadRejectsInvalidJSON(t *testing.T) {
	if err := ValidatePayload(`{"ok":true}`); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePayload(`not-json`); err == nil {
		t.Fatal("expected invalid JSON to be rejected")
	}
}

func TestPendingRejectsInvalidPayload(t *testing.T) {
	db, mock, gormDB := newOutboxMock(t)
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, tema, contenido::text FROM erp_v4.gs_evento_salida")).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "tema", "contenido"}).AddRow(1, "venta", "not-json"))
	if _, err := (PostgresStore{DB: gormDB}).Pending(context.Background(), 1); err == nil {
		t.Fatal("expected invalid payload error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProcessPendingPublishesBeforeAcknowledging(t *testing.T) {
	db, mock, gormDB := newOutboxMock(t)
	defer db.Close()
	p := &publisher{}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, tema, contenido::text FROM erp_v4.gs_evento_salida")).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "tema", "contenido"}).AddRow(9, "venta.confirmada", `{"id":9}`))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE erp_v4.gs_evento_salida SET publicado_en=now()")).
		WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := (PostgresStore{DB: gormDB}).ProcessPending(context.Background(), 1, p); err != nil {
		t.Fatal(err)
	}
	if p.calls != 1 {
		t.Fatalf("expected one publication, got %d", p.calls)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMarkFailedPersistsRetryDiagnostic(t *testing.T) {
	db, mock, gormDB := newOutboxMock(t)
	defer db.Close()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE erp_v4.gs_evento_salida SET reintentos=reintentos+1, publicado_error=$2")).
		WithArgs(int64(7), "temporary").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := (PostgresStore{DB: gormDB}).MarkFailed(context.Background(), "7", errors.New("temporary")); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProcessPendingRecordsPublishFailureAndLeavesPending(t *testing.T) {
	db, mock, gormDB := newOutboxMock(t)
	defer db.Close()
	p := &publisher{failures: 1}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, tema, contenido::text FROM erp_v4.gs_evento_salida")).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "tema", "contenido"}).AddRow(9, "venta.confirmada", `{"id":9}`))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE erp_v4.gs_evento_salida SET reintentos=reintentos+1, publicado_error=$2")).
		WithArgs(int64(9), "temporary publish failure").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := (PostgresStore{DB: gormDB}).ProcessPending(context.Background(), 1, p); err != ErrTemporary {
		t.Fatalf("expected temporary failure, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
