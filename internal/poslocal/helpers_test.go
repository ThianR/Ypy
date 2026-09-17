package poslocal

import (
	"database/sql"
	"gorm.io/gorm"
)

func CloseAgentDB(db *gorm.DB) {
	if db != nil {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
}

func agentRow(db *gorm.DB, query string, args ...any) (*sql.Row, error) {
	result := db.Raw(query, args...)
	return result.Row(), result.Error
}
