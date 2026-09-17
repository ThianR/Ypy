package poscentral

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newPosMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *gorm.DB) {
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

func TestPostgresStoreApplyIsIdempotent(t *testing.T) {
	db, mock, gormDB := newPosMock(t)
	defer db.Close()
	op := Operation{ID: "00000000-0000-0000-0000-000000000001", EmpresaID: "1", TerminalID: "2", Payload: `{"sequence":1,"item":"A"}`}
	insert := regexp.QuoteMeta("INSERT INTO erp_v4.gp_evento_entrada")
	mock.ExpectExec(insert).WithArgs(int64(1), sqlmock.AnyArg(), int64(2), int64(1), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT empresa_id, terminal_id, secuencia, contenido_hash FROM erp_v4.gp_evento_entrada WHERE uid=$1")).WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"empresa_id", "terminal_id", "secuencia", "contenido_hash"}).AddRow(1, 2, 1, hash(op.Payload)))
	if _, err := (PostgresStore{DB: gormDB}).Apply(op); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresStoreEnqueuesValidatedAcceptedEvent(t *testing.T) {
	db, mock, gormDB := newPosMock(t)
	defer db.Close()
	op := Operation{ID: "00000000-0000-0000-0000-000000000007", EmpresaID: "1", TerminalID: "2", Payload: `{"sequence":1,"item":"A"}`}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO erp_v4.gs_evento_salida(empresa_id,uid,tema,contenido)")).WithArgs(int64(1), sqlmock.AnyArg(), "pos.operation.accepted", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	if err := (PostgresStore{DB: gormDB}).EnqueueAccepted(context.Background(), op, time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresStoreApplyDetectsConflict(t *testing.T) {
	db, mock, gormDB := newPosMock(t)
	defer db.Close()
	op := Operation{ID: "00000000-0000-0000-0000-000000000002", EmpresaID: "1", TerminalID: "2", Payload: `{"sequence":1,"item":"A"}`}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO erp_v4.gp_evento_entrada")).WithArgs(int64(1), sqlmock.AnyArg(), int64(2), int64(1), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT empresa_id, terminal_id, secuencia, contenido_hash FROM erp_v4.gp_evento_entrada WHERE uid=$1")).WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"empresa_id", "terminal_id", "secuencia", "contenido_hash"}).AddRow(1, 2, 1, "different"))
	if _, err := (PostgresStore{DB: gormDB}).Apply(op); err != ErrConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresStoreSavesCursorMonotonically(t *testing.T) {
	db, mock, gormDB := newPosMock(t)
	defer db.Close()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO erp_v4.gp_sync_cursor")).WithArgs(int64(1), int64(2), int64(8)).WillReturnResult(sqlmock.NewResult(1, 1))
	if err := (PostgresStore{DB: gormDB}).SaveCursor(context.Background(), 1, 2, 8); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresStoreReadsMissingCursorAsZero(t *testing.T) {
	db, mock, gormDB := newPosMock(t)
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT cursor FROM erp_v4.gp_sync_cursor WHERE empresa_id=$1 AND terminal_id=$2")).WithArgs(int64(1), int64(2)).WillReturnError(sql.ErrNoRows)
	cursor, err := (PostgresStore{DB: gormDB}).Cursor(context.Background(), 1, 2)
	if err != nil || cursor != 0 {
		t.Fatalf("unexpected missing cursor: %d %v", cursor, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresStoreRejectsInvalidInboxSequence(t *testing.T) {
	db, _, gormDB := newPosMock(t)
	defer db.Close()
	op := Operation{ID: "00000000-0000-0000-0000-000000000003", EmpresaID: "1", TerminalID: "2", Payload: `{"item":"A"}`}
	if _, err := (PostgresStore{DB: gormDB}).Apply(op); err != ErrInvalidOperation {
		t.Fatalf("expected invalid operation, got %v", err)
	}
}

func TestPostgresStoreRejectsInvalidCursorScope(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, scope := range [][2]int64{{0, 1}, {1, 0}, {-1, 1}} {
		if err := (PostgresStore{}).SaveCursor(context.Background(), scope[0], scope[1], 0); err == nil {
			t.Fatalf("expected invalid scope rejection for %v", scope)
		}
		if _, err := (PostgresStore{}).Cursor(context.Background(), scope[0], scope[1]); err == nil {
			t.Fatalf("expected invalid cursor read scope rejection for %v", scope)
		}
	}
}

func TestValidateSalePayloadRequiresCommercialFields(t *testing.T) {
	valid := `{"sequence":1,"occurred_at":"2026-01-01T10:00:00Z","permiso_id":1,"usuario_id":2,"talonario_id":3,"rango_id":4,"numero":5,"tipo_id":6,"moneda_id":7,"lineas":[{"articulo_id":8}]}`
	if err := validateSalePayload(valid); err != nil {
		t.Fatalf("valid sale rejected: %v", err)
	}
	if err := validateSalePayload(`{"sequence":1,"lineas":[]}`); err != ErrInvalidOperation {
		t.Fatalf("expected invalid sale payload, got %v", err)
	}
}

func TestPostgresStoreIntegratesSaleWithReceipt(t *testing.T) {
	db, mock, gormDB := newPosMock(t)
	defer db.Close()
	op := Operation{ID: "00000000-0000-0000-0000-000000000014", EmpresaID: "1", TerminalID: "2", Payload: `{"sequence":2,"occurred_at":"2026-09-16T10:00:00Z","permiso_id":1,"usuario_id":2,"talonario_id":3,"rango_id":4,"numero":5,"tipo_id":6,"moneda_id":7,"cliente_id":8,"apertura_id":9,"total":"100.000000","formas":[{"medio_id":10,"importe":"100.000000"}],"lineas":[{"articulo_id":11}]}`}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT erp_v4.gp_integrar_venta_cobro($1,$2,$3,$4,$5,$6::jsonb)")).WithArgs(int64(1), int64(2), int64(2), sqlmock.AnyArg(), time.Date(2026, 9, 16, 10, 0, 0, 0, time.UTC), op.Payload).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(90))
	got, err := (PostgresStore{DB: gormDB}).IntegrateSaleWithReceipt(context.Background(), op)
	if err != nil || got != 90 {
		t.Fatalf("sale=%d err=%v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
