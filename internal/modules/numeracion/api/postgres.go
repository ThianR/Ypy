package api

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

var ErrInvalidRequest = errors.New("invalid numbering request")

// PostgresStore delega las decisiones de numeración a las funciones V4.
// Esto mantiene la exclusividad de rangos y permisos en la misma transacción.
type PostgresStore struct{ DB *gorm.DB }

func numberingResult(db *gorm.DB, query string, args ...any) (int64, error) {
	var id int64
	result := db.Raw(query, args...)
	row := result.Row()
	if result.Error != nil {
		return 0, result.Error
	}
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *PostgresStore) Reserve(ctx context.Context, companyID, bookletID, terminalID, userID, size int64, key uuid.UUID, reserve bool) (int64, error) {
	if s == nil || s.DB == nil || companyID <= 0 || bookletID <= 0 || terminalID <= 0 || userID <= 0 || size <= 0 || key == uuid.Nil {
		return 0, ErrInvalidRequest
	}
	return numberingResult(s.DB.WithContext(ctx), `SELECT gs_reservar_rango($1,$2,$3,$4,$5,$6,$7)`, companyID, bookletID, terminalID, userID, size, key, reserve)
}

func (s *PostgresStore) Assign(ctx context.Context, companyID, bookletID, userID, terminalID, rangeID, requested int64, key uuid.UUID, at time.Time) (int64, error) {
	if s == nil || s.DB == nil || companyID <= 0 || bookletID <= 0 || userID <= 0 || terminalID <= 0 || key == uuid.Nil {
		return 0, ErrInvalidRequest
	}
	if requested <= 0 && rangeID <= 0 {
		return 0, ErrInvalidRequest
	}
	if rangeID > 0 {
		return numberingResult(s.DB.WithContext(ctx), `SELECT gs_asignar_numero($1,$2,$3,$4,$5,$6,$7,$8)`, companyID, bookletID, userID, key, terminalID, rangeID, requested, at)
	} else {
		return numberingResult(s.DB.WithContext(ctx), `SELECT gs_asignar_numero($1,$2,$3,$4,$5,NULL,$6,$7)`, companyID, bookletID, userID, key, terminalID, requested, at)
	}
}
