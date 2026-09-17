package api

import (
	"context"
	"errors"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrInvalidReturnRequest = errors.New("invalid return request")

type ReturnStore struct{ DB *gorm.DB }

type CreditNote struct {
	ID         uuid.UUID `json:"id"`
	CompanyID  int64     `json:"empresa_id"`
	BranchID   int64     `json:"sucursal_id"`
	UserID     int64     `json:"usuario_id"`
	TypeID     int64     `json:"tipo_id"`
	CustomerID int64     `json:"cliente_id"`
	CurrencyID int64     `json:"moneda_id"`
	OriginID   int64     `json:"comprobante_origen_id"`
	Total      string    `json:"total"`
	Date       string    `json:"fecha"`
	Reason     string    `json:"motivo_excepcion"`
}

type CreditNoteStore struct{ DB *gorm.DB }

func (s *CreditNoteStore) Create(ctx context.Context, n CreditNote) (int64, error) {
	amount, valid := new(big.Rat).SetString(n.Total)
	if s == nil || s.DB == nil || n.ID == uuid.Nil || n.CompanyID <= 0 || n.BranchID <= 0 || n.UserID <= 0 || n.TypeID <= 0 || n.CustomerID <= 0 || n.CurrencyID <= 0 || n.OriginID <= 0 || !valid || amount.Sign() < 0 || strings.TrimSpace(n.Reason) == "" || n.Date == "" {
		return 0, ErrInvalidReturnRequest
	}
	var id int64
	query := s.DB.WithContext(ctx).Raw(`INSERT INTO gv_nota_credito_cabecera(empresa_id,uid,sucursal_id,usuario_id,tipo_id,tercero_id,moneda_id,fecha_operacion,total,motivo_excepcion,comprobante_origen_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id`, n.CompanyID, n.ID, n.BranchID, n.UserID, n.TypeID, n.CustomerID, n.CurrencyID, n.Date, n.Total, strings.TrimSpace(n.Reason), n.OriginID)
	row := query.Row()
	err := query.Error
	if err == nil {
		err = row.Scan(&id)
	}
	return id, err
}

func (s *CreditNoteStore) Confirm(ctx context.Context, companyID, noteID int64) error {
	if s == nil || s.DB == nil || companyID <= 0 || noteID <= 0 {
		return ErrInvalidReturnRequest
	}
	return s.DB.WithContext(ctx).Exec(`SELECT gs_confirmar($1,$2,$3)`, companyID, "gv_nota_credito", noteID).Error
}

func (s *ReturnStore) Create(ctx context.Context, companyID, lineID int64, quantity, reason string, id uuid.UUID) (int64, error) {
	amount, valid := new(big.Rat).SetString(quantity)
	if s == nil || s.DB == nil || companyID <= 0 || lineID <= 0 || id == uuid.Nil || !valid || amount.Sign() <= 0 || strings.TrimSpace(reason) == "" {
		return 0, ErrInvalidReturnRequest
	}
	var returnID int64
	query := s.DB.WithContext(ctx).Raw(`SELECT erp_v4.gv_devolver($1,$2,$3,$4,$5)`, companyID, lineID, quantity, id, strings.TrimSpace(reason))
	row := query.Row()
	err := query.Error
	if err == nil {
		err = row.Scan(&returnID)
	}
	return returnID, err
}
