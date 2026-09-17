package poslocal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSQLiteIntegrityCheck(t *testing.T) {
	db, err := OpenStore(filepath.Join(t.TempDir(), "pos.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer CloseAgentDB(db)
	if err := CheckIntegrity(db); err != nil {
		t.Fatal(err)
	}
}

func TestBackupCanBeOpenedAndVerified(t *testing.T) {
	db, err := OpenStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer CloseAgentDB(db)
	destination := filepath.Join(t.TempDir(), "pos-backup.db")
	if err := BackupDatabase(db, destination); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(destination); err != nil {
		t.Fatal(err)
	}
}

func TestRestoreCreatesVerifiedIndependentCopy(t *testing.T) {
	source := filepath.Join(t.TempDir(), "source.db")
	db, err := OpenStore(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := SaveOperation(db, "restore-1", "1001", "25", "PYG", "PENDING", "2026-09-16"); err != nil {
		t.Fatal(err)
	}
	CloseAgentDB(db)
	destination := filepath.Join(t.TempDir(), "restored.db")
	if err := restoreDatabase(source, destination); err != nil {
		t.Fatal(err)
	}
	restored, err := OpenStore(destination)
	if err != nil {
		t.Fatal(err)
	}
	defer CloseAgentDB(restored)
	var count int
	row, queryErr := agentRow(restored, "SELECT count(*) FROM pos_operations WHERE operation_id='restore-1'")
	if err := queryErr; err != nil || row.Scan(&count) != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected restored operation, got %d", count)
	}
}

func TestBackupAndRestoreNeverOverwriteExistingDestination(t *testing.T) {
	db, err := OpenStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer CloseAgentDB(db)
	destination := filepath.Join(t.TempDir(), "existing.db")
	if err := os.WriteFile(destination, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := BackupDatabase(db, destination); err == nil {
		t.Fatal("backup overwrote existing destination")
	}
	if err := restoreDatabase(destination, filepath.Join(t.TempDir(), "restored.db")); err == nil {
		t.Fatal("invalid source was accepted")
	}
	contents, err := os.ReadFile(destination)
	if err != nil || string(contents) != "keep" {
		t.Fatal("existing destination was modified")
	}
}
