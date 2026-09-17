package poslocal

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestPersistentPrinterSurvivesReopenAndRetries(t *testing.T) {
	path := t.TempDir() + "/pos.db"
	db, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	CloseAgentDB(db)
	gormDB, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	job, err := (PersistentPrinter{DB: gormDB}).Enqueue("sale-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := (PersistentPrinter{DB: gormDB}).Print(job.ID, false); err != ErrPrinterUnavailable {
		t.Fatalf("expected printer error, got %v", err)
	}
	CloseAgentDB(db)
	if sqlDB, closeErr := gormDB.DB(); closeErr == nil {
		_ = sqlDB.Close()
	}
	db, err = OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	defer CloseAgentDB(db)
	gormDB, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if sqlDB, closeErr := gormDB.DB(); closeErr == nil {
			_ = sqlDB.Close()
		}
	}()
	printer := PersistentPrinter{DB: gormDB}
	got, err := printer.Job(job.ID)
	if err != nil || got.Status != "UNKNOWN" || got.Attempts != 1 {
		t.Fatalf("unexpected persisted job: %#v, %v", got, err)
	}
	pending, err := printer.Pending()
	if err != nil || len(pending) != 1 || pending[0].Status != "UNKNOWN" {
		t.Fatalf("expected UNKNOWN job in recovery queue: %#v, %v", pending, err)
	}
	if err := printer.Print(job.ID, true); err != nil {
		t.Fatal(err)
	}
	got, err = printer.Job(job.ID)
	if err != nil || got.Status != "PRINTED" || got.Attempts != 2 {
		t.Fatalf("unexpected retry: %#v, %v", got, err)
	}
	pending, err = printer.Pending()
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 0 {
		t.Fatalf("expected no pending jobs, got %d", len(pending))
	}
}

func TestPersistentPrinterRejectsNilStore(t *testing.T) {
	p := PersistentPrinter{}
	if _, err := p.Pending(); err != ErrInvalidPrinterStore {
		t.Fatalf("pending accepted nil store: %v", err)
	}
	if _, err := p.Job("ticket-x"); err != ErrInvalidPrinterStore {
		t.Fatalf("job accepted nil store: %v", err)
	}
	if err := p.Print("ticket-x", true); err != ErrInvalidPrinterStore {
		t.Fatalf("print accepted nil store: %v", err)
	}
}
