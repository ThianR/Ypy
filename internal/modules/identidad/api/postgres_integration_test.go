package api

import (
	"context"
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestSessionStoreAgainstPostgreSQL(t *testing.T) {
	dsn := os.Getenv("YPY_DATABASE_URL")
	if dsn == "" {
		t.Skip("YPY_DATABASE_URL no definido")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	if err := sqlDB.PingContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	sessionID, err := (SessionStore{DB: db}).Authenticate(context.Background(), "demo", "demo123", 1, 1, 1, time.Now().UTC(), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (SessionStore{DB: db}).LoadContext(context.Background(), sessionID, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
}
