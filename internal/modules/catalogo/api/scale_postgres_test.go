package api

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestLoadV4ScaleRulesFiltersInvalidRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT prefijo AS prefix,longitud AS length,inicio_producto AS product_start,largo_producto AS product_length,inicio_valor AS value_start,largo_valor AS value_length,decimales,contenido AS content FROM erp_v4.gp_regla_balanza")).
		WithArgs(int64(4)).WillReturnRows(sqlmock.NewRows([]string{"prefix", "length", "product_start", "product_length", "value_start", "value_length", "decimales", "content"}).
		AddRow("99", 13, 3, 5, 8, 6, 2, "PRECIO").
		AddRow("bad", 0, 1, 1, 1, 1, 0, "PESO"))
	rules, err := LoadV4ScaleRules(context.Background(), gormDB, 4)
	if err != nil || len(rules) != 1 || rules[0].Prefix != "99" {
		t.Fatalf("rules=%#v err=%v", rules, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestLoadV4ScaleRulesRejectsInvalidStore(t *testing.T) {
	if _, err := LoadV4ScaleRules(context.Background(), nil, 1); err == nil {
		t.Fatal("expected invalid store error")
	}
	if _, err := LoadV4ScaleRules(context.Background(), nil, 0); err == nil {
		t.Fatal("expected invalid company error")
	}
}

func TestLoadV4ScaleRulesPropagatesQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT prefijo AS prefix,longitud AS length,inicio_producto AS product_start,largo_producto AS product_length,inicio_valor AS value_start,largo_valor AS value_length,decimales,contenido AS content FROM erp_v4.gp_regla_balanza")).
		WithArgs(int64(4)).WillReturnError(errors.New("database unavailable"))
	if _, err := LoadV4ScaleRules(context.Background(), gormDB, 4); err == nil || err.Error() != "database unavailable" {
		t.Fatalf("expected database error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
