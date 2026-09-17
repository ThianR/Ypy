package api

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestCreditStoreAppliesCreditThroughV4(t *testing.T) {
	db, mock, gormDB := newTreasuryMock(t)
	defer db.Close()
	id := uuid.MustParse("00000000-0000-0000-0000-000000000018")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT erp_v4.gf_usar_credito($1,$2,$3,$4,$5,$6)")).WithArgs(int64(4), "gv", int64(80), int64(90), "25.000000", id).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(94))
	got, err := (&CreditStore{DB: gormDB}).Apply(context.Background(), 4, 80, 90, "25.000000", id)
	if err != nil || got != 94 {
		t.Fatalf("application=%d err=%v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
