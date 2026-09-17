package poslocal

import (
	"database/sql"
	"errors"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"os"
	"path/filepath"
)

func BackupDatabase(db *gorm.DB, destination string) error {
	if db == nil || destination == "" {
		return errors.New("invalid backup target")
	}
	if _, err := os.Stat(destination); err == nil {
		return errors.New("backup destination already exists")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if err := CheckIntegrityDB(sqlDB); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	if _, err := sqlDB.Exec("VACUUM INTO ?", destination); err != nil {
		return err
	}
	backup, err := sql.Open("sqlite", destination)
	if err != nil {
		return err
	}
	defer backup.Close()
	return CheckIntegrityDB(backup)
}

func restoreDatabase(source, destination string) error {
	if source == "" || destination == "" {
		return errors.New("invalid restore target")
	}
	if _, err := os.Stat(destination); err == nil {
		return errors.New("restore destination already exists")
	}
	sourceDB, err := gorm.Open(sqlite.Open(source), &gorm.Config{})
	if err != nil {
		return err
	}
	defer func() {
		if db, closeErr := sourceDB.DB(); closeErr == nil {
			_ = db.Close()
		}
	}()
	return BackupDatabase(sourceDB, destination)
}

func CheckIntegrityDB(db *sql.DB) error {
	var result string
	if err := db.QueryRow("PRAGMA integrity_check").Scan(&result); err != nil {
		return err
	}
	if result != "ok" {
		return errors.New("sqlite integrity check failed: " + result)
	}
	return nil
}
