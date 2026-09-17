package poslocal

import (
	"database/sql"
	"encoding/json"
	"errors"

	databaseplatform "github.com/ypy-erp/ypy/internal/platform/database"
	"gorm.io/gorm"
)

var ErrOperationConflict = errors.New("operation content conflict")
var ErrInvalidOperation = errors.New("invalid local operation")

func OpenStore(path string) (*gorm.DB, error) {
	db, err := databaseplatform.OpenSQLite(path, databaseplatform.DefaultConfig())
	if err != nil {
		return nil, err
	}
	err = db.Exec(`CREATE TABLE IF NOT EXISTS pos_operations (
		operation_id TEXT PRIMARY KEY, number TEXT NOT NULL, total TEXT NOT NULL,
		currency TEXT NOT NULL, state TEXT NOT NULL CHECK(state IN ('PENDING','SENDING','APPLIED','CONFLICT')), created_at TEXT NOT NULL,
		CHECK(operation_id <> ''), CHECK(number <> ''), CHECK(total <> ''), CHECK(currency <> ''), CHECK(created_at <> '')
	)`).Error
	if err == nil {
		err = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_pos_operations_number ON pos_operations(number)`).Error
	}
	if err == nil {
		err = db.Exec(`CREATE TABLE IF NOT EXISTS pos_print_jobs (
			job_id TEXT PRIMARY KEY, operation_id TEXT NOT NULL, status TEXT NOT NULL,
			attempts INTEGER NOT NULL DEFAULT 0, UNIQUE(operation_id)
		)`).Error
	}
	if err == nil {
		err = db.Exec(`CREATE TABLE IF NOT EXISTS pos_scale_rules (
			empresa_id TEXT PRIMARY KEY CHECK(empresa_id <> ''), rules_json TEXT NOT NULL CHECK(rules_json <> ''), updated_at TEXT NOT NULL CHECK(updated_at <> '')
		)`).Error
	}
	if err != nil {
		_ = databaseplatform.Close(db)
		return nil, err
	}
	if result := db.Exec(`UPDATE pos_operations SET state='PENDING' WHERE state='SENDING'`); result.Error != nil {
		_ = databaseplatform.Close(db)
		return nil, result.Error
	}
	return db, nil
}

func SaveScaleRules(db *gorm.DB, empresaID, rulesJSON, updatedAt string) error {
	if db == nil || empresaID == "" || rulesJSON == "" || updatedAt == "" || !ValidScaleRulesJSON(rulesJSON) {
		return ErrInvalidOperation
	}
	return db.Exec(`INSERT INTO pos_scale_rules(empresa_id,rules_json,updated_at) VALUES(?,?,?) ON CONFLICT(empresa_id) DO UPDATE SET rules_json=excluded.rules_json, updated_at=excluded.updated_at`, empresaID, rulesJSON, updatedAt).Error
}

func ValidScaleRulesJSON(rulesJSON string) bool {
	var rules []json.RawMessage
	if !json.Valid([]byte(rulesJSON)) || json.Unmarshal([]byte(rulesJSON), &rules) != nil {
		return false
	}
	for _, rule := range rules {
		var object struct {
			Prefix        string `json:"prefix"`
			Length        int    `json:"length"`
			ProductStart  int    `json:"product_start"`
			ProductLength int    `json:"product_length"`
			ValueStart    int    `json:"value_start"`
			ValueLength   int    `json:"value_length"`
			Decimals      int    `json:"decimals"`
			Content       string `json:"content"`
		}
		if json.Unmarshal(rule, &object) != nil || object.Prefix == "" || object.Length <= 0 || len(object.Prefix) > object.Length || object.ProductStart <= 0 || object.ProductLength <= 0 || object.ValueStart <= 0 || object.ValueLength <= 0 || object.Decimals < 0 || object.Decimals > 6 || object.ProductStart+object.ProductLength-1 > object.Length || object.ValueStart+object.ValueLength-1 > object.Length {
			return false
		}
		if object.Content != "PESO" && object.Content != "PRECIO" {
			return false
		}
	}
	return true
}

func LoadScaleRules(db *gorm.DB, empresaID string) (string, error) {
	if db == nil || empresaID == "" {
		return "", ErrInvalidOperation
	}
	var rules string
	if result := db.Raw(`SELECT rules_json FROM pos_scale_rules WHERE empresa_id=?`, empresaID).Scan(&rules); result.Error != nil {
		return "", result.Error
	}
	if !ValidScaleRulesJSON(rules) {
		return "", ErrInvalidOperation
	}
	return rules, nil
}

func SaveOperation(db *gorm.DB, operationID, number, total, currency, state, createdAt string) error {
	if db == nil || operationID == "" || number == "" || total == "" || currency == "" || createdAt == "" {
		return ErrInvalidOperation
	}
	switch state {
	case "PENDING", "SENDING", "APPLIED", "CONFLICT":
	default:
		return ErrInvalidOperation
	}
	var oldTotal string
	result := db.Raw(`SELECT total FROM pos_operations WHERE operation_id = ?`, operationID).Scan(&oldTotal)
	err := result.Error
	if err == nil && result.RowsAffected > 0 {
		if oldTotal != total {
			return ErrOperationConflict
		}
		return nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	var existingID string
	if result := db.Raw(`SELECT operation_id FROM pos_operations WHERE number = ?`, number).Scan(&existingID); result.Error == nil && result.RowsAffected > 0 {
		return ErrOperationConflict
	} else if result.Error != nil && !errors.Is(result.Error, sql.ErrNoRows) && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}
	return db.Exec(`INSERT INTO pos_operations(operation_id,number,total,currency,state,created_at) VALUES(?,?,?,?,?,?)`, operationID, number, total, currency, state, createdAt).Error
}

func CheckIntegrity(db *gorm.DB) error {
	var result string
	if query := db.Raw(`PRAGMA integrity_check`).Scan(&result); query.Error != nil {
		return query.Error
	}
	if result != "ok" {
		return errors.New("sqlite integrity check failed: " + result)
	}
	return nil
}
