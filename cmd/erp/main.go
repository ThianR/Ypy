package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ypy-erp/ypy/internal/bootstrap"
	databaseplatform "github.com/ypy-erp/ypy/internal/platform/database"
	"gorm.io/gorm"
)

func main() {
	var db *sql.DB
	var gormDB *gorm.DB
	if dsn := os.Getenv("YPY_DATABASE_URL"); dsn != "" {
		connection, err := databaseplatform.OpenPostgres(dsn, databaseplatform.DefaultConfig())
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		gormDB = connection
		db, err = databaseplatform.SQLDB(connection)
		if err != nil {
			log.Fatalf("Failed to get raw DB: %v", err)
		}
	}
	defer func() {
		if gormDB != nil {
			_ = databaseplatform.Close(gormDB)
		}
	}()

	handler := bootstrap.BuildRouter(db, gormDB)

	fmt.Println("Ypy ERP listening on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}
