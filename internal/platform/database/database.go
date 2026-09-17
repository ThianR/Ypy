package database

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config contiene la configuración común de las conexiones de persistencia.
type Config struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultConfig devuelve una configuración conservadora para el MVP.
func DefaultConfig() Config {
	return Config{MaxOpenConns: 10, MaxIdleConns: 5, ConnMaxLifetime: time.Hour, ConnMaxIdleTime: 15 * time.Minute}
}

// OpenPostgres abre una conexión GORM sobre PostgreSQL y configura su pool.
func OpenPostgres(dsn string, cfg Config) (*gorm.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("dsn de PostgreSQL vacío")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn), TranslateError: true})
	if err != nil {
		return nil, err
	}
	if err := configurePool(db, cfg); err != nil {
		return nil, err
	}
	return db, nil
}

// OpenSQLite abre una conexión GORM sobre SQLite sin requerir CGO.
func OpenSQLite(path string, cfg Config) (*gorm.DB, error) {
	if path == "" {
		return nil, fmt.Errorf("ruta de SQLite vacía")
	}
	// SQLite en memoria requiere una única conexión para conservar el mismo
	// esquema y los mismos datos entre operaciones consecutivas.
	if strings.HasPrefix(path, ":memory:") {
		cfg.MaxOpenConns = 1
		cfg.MaxIdleConns = 1
	}
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn), TranslateError: true})
	if err != nil {
		return nil, err
	}
	if err := configurePool(db, cfg); err != nil {
		return nil, err
	}
	return db, nil
}

func configurePool(db *gorm.DB, cfg Config) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}
	return nil
}

// Close libera el pool subyacente administrado por GORM.
func Close(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// EnvInt permite configurar límites del pool sin duplicar análisis de entorno.
func EnvInt(name string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

// SQLDB expone el pool subyacente durante la migración gradual de adaptadores.
func SQLDB(db *gorm.DB) (*sql.DB, error) {
	if db == nil {
		return nil, fmt.Errorf("conexión GORM nula")
	}
	return db.DB()
}
