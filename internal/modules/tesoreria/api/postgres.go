package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"math/big"
	"regexp"
	"strings"
)

var ErrInvalidCashScope = errors.New("invalid cash scope")
var cashAmount = regexp.MustCompile(`^(0|[1-9][0-9]*)(\.[0-9]+)?$`)

type PostgresStore struct{ DB *gorm.DB }

var ErrCashMovementConflict = errors.New("cash movement UUID content conflict")

type CashMovement struct {
	ID                             uuid.UUID
	CompanyID, OpeningID, MethodID int64
	Kind, Amount, Reason, Date     string
}

func (s *PostgresStore) AddMovement(ctx context.Context, m CashMovement) (int64, error) {
	parsedAmount, validAmount := new(big.Rat).SetString(m.Amount)
	if s == nil || s.DB == nil || m.ID == uuid.Nil || m.CompanyID <= 0 || m.OpeningID <= 0 || m.MethodID <= 0 || (m.Kind != "INGRESO" && m.Kind != "RETIRO") || !validAmount || !cashAmount.MatchString(m.Amount) || parsedAmount.Sign() <= 0 || strings.TrimSpace(m.Reason) == "" || m.Date == "" {
		return 0, ErrInvalidCashScope
	}
	var id, company, opening, method int64
	var kind, amount, reason, date string
	query := s.DB.WithContext(ctx).Raw(`SELECT id,empresa_id,apertura_id,medio_id,tipo,importe::text,motivo,fecha::text FROM erp_v4.gf_movimiento_caja WHERE uid=$1`, m.ID)
	row := query.Row()
	err := query.Error
	if err == nil {
		err = row.Scan(&id, &company, &opening, &method, &kind, &amount, &reason, &date)
	}
	if err == nil {
		if company != m.CompanyID || opening != m.OpeningID || method != m.MethodID || kind != m.Kind || amount != m.Amount || reason != strings.TrimSpace(m.Reason) || date != m.Date {
			return 0, ErrCashMovementConflict
		}
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	query = s.DB.WithContext(ctx).Raw(`INSERT INTO erp_v4.gf_movimiento_caja(empresa_id,uid,apertura_id,medio_id,tipo,importe,motivo,fecha) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`, m.CompanyID, m.ID, m.OpeningID, m.MethodID, m.Kind, m.Amount, strings.TrimSpace(m.Reason), m.Date)
	row = query.Row()
	err = query.Error
	if err == nil {
		err = row.Scan(&id)
	}
	return id, err
}

func (s *PostgresStore) Open(ctx context.Context, companyID, cashID, userID, terminalID int64, fund string) (int64, error) {
	if s == nil || s.DB == nil || companyID <= 0 || cashID <= 0 || userID <= 0 || terminalID <= 0 || !cashAmount.MatchString(fund) {
		return 0, ErrInvalidCashScope
	}
	var openingID int64
	query := s.DB.WithContext(ctx).Raw(`INSERT INTO erp_v4.gf_caja_apertura(empresa_id,caja_id,usuario_id,terminal_id,fondo) VALUES($1,$2,$3,$4,$5) RETURNING id`, companyID, cashID, userID, terminalID, fund)
	row := query.Row()
	err := query.Error
	if err == nil {
		err = row.Scan(&openingID)
	}
	return openingID, err
}

func (s *PostgresStore) Close(ctx context.Context, companyID, openingID int64, counts []byte) error {
	if s == nil || s.DB == nil || companyID <= 0 || openingID <= 0 || !json.Valid(counts) {
		return ErrInvalidCashScope
	}
	return s.DB.WithContext(ctx).Exec(`SELECT erp_v4.gf_cerrar_caja($1,$2,$3::jsonb)`, companyID, openingID, string(counts)).Error
}
