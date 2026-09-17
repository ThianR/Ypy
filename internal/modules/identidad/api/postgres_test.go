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

func newIdentityMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *gorm.DB) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return sqlDB, mock, gormDB
}

func TestSessionStoreLoadsActiveScopedContext(t *testing.T) {
	db, mock, gormDB := newIdentityMock(t)
	defer db.Close()
	now := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT usuario_id AS user_id,empresa_id AS company_id,sucursal_id AS branch_id,terminal_id,expira_en AS expires_at,revocada_en AS revoked_at FROM erp_v4.ypy_usuario_sesion")).WithArgs("00000000-0000-0000-0000-000000000001").WillReturnRows(sqlmock.NewRows([]string{"user_id", "company_id", "branch_id", "terminal_id", "expires_at", "revoked_at"}).AddRow(7, 4, 2, 9, now.Add(time.Hour), nil))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).WithArgs(int64(7), int64(4), int64(2), int64(9)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT p.codigo")).WithArgs(int64(4), int64(7), int64(2)).WillReturnRows(sqlmock.NewRows([]string{"codigo"}).AddRow("sale.confirm"))
	ctx, err := (SessionStore{DB: gormDB}).LoadContext(context.Background(), "00000000-0000-0000-0000-000000000001", now)
	if err != nil || !ctx.Active || ctx.UserID != "7" || ctx.EmpresaID != "4" || ctx.TerminalID != "9" {
		t.Fatalf("ctx=%#v err=%v", ctx, err)
	}
	if err := ctx.Can("sale.confirm", "4", "2"); err != nil {
		t.Fatalf("expected role permission, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSessionStoreRejectsExpiredOrIncompleteSession(t *testing.T) {
	db, mock, gormDB := newIdentityMock(t)
	defer db.Close()
	now := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT usuario_id AS user_id,empresa_id AS company_id,sucursal_id AS branch_id,terminal_id,expira_en AS expires_at,revocada_en AS revoked_at FROM erp_v4.ypy_usuario_sesion")).WithArgs("00000000-0000-0000-0000-000000000002").WillReturnRows(sqlmock.NewRows([]string{"user_id", "company_id", "branch_id", "terminal_id", "expires_at", "revoked_at"}).AddRow(7, 4, 2, nil, now.Add(time.Hour), nil))
	if _, err := (SessionStore{DB: gormDB}).LoadContext(context.Background(), "00000000-0000-0000-0000-000000000002", now); err != ErrSessionExpired {
		t.Fatalf("expected expired/incomplete session, got %v", err)
	}
}

func TestSessionStoreCreatesScopedSession(t *testing.T) {
	db, mock, gormDB := newIdentityMock(t)
	defer db.Close()
	now := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO erp_v4.ypy_usuario_sesion")).WithArgs(sqlmock.AnyArg(), int64(7), int64(4), int64(2), int64(9), now, now.Add(time.Hour)).WillReturnResult(sqlmock.NewResult(0, 1))
	id, err := (SessionStore{DB: gormDB}).Create(context.Background(), 7, 4, 2, 9, now, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := uuid.Parse(id); err != nil {
		t.Fatalf("invalid created session id: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSessionStoreAuthenticatesActiveScopedUser(t *testing.T) {
	db, mock, gormDB := newIdentityMock(t)
	defer db.Close()
	now := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.id FROM erp_v4.gs_usuario u")).WithArgs("cashier", "secret", int64(4), int64(2), int64(9)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO erp_v4.ypy_usuario_sesion")).WithArgs(sqlmock.AnyArg(), int64(7), int64(4), int64(2), int64(9), now, now.Add(time.Hour)).WillReturnResult(sqlmock.NewResult(0, 1))
	id, err := (SessionStore{DB: gormDB}).Authenticate(context.Background(), "cashier", "secret", 4, 2, 9, now, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := uuid.Parse(id); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSessionStoreRevokesSessionIdempotently(t *testing.T) {
	db, mock, gormDB := newIdentityMock(t)
	defer db.Close()
	at := time.Date(2026, 9, 16, 13, 0, 0, 0, time.UTC)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE erp_v4.ypy_usuario_sesion SET revocada_en=$2")).WithArgs("00000000-0000-0000-0000-000000000003", at).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := (SessionStore{DB: gormDB}).Revoke(context.Background(), "00000000-0000-0000-0000-000000000003", at); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
