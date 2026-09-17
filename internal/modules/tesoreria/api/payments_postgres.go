package api

import (
	"context"
	"errors"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrInvalidPaymentRequest = errors.New("invalid payment request")

type PaymentDocument struct {
	ID         uuid.UUID `json:"id"`
	CompanyID  int64     `json:"empresa_id"`
	ProviderID int64     `json:"proveedor_id"`
	OpeningID  int64     `json:"apertura_id"`
	CurrencyID int64     `json:"moneda_id"`
	Amount     string    `json:"importe"`
	Date       string    `json:"fecha"`
}

type PaymentStore struct{ DB *gorm.DB }

type PaymentForm struct {
	MethodID  int64  `json:"medio_id"`
	Amount    string `json:"importe"`
	Reference string `json:"referencia"`
}

func (s *PaymentStore) CreateAndConfirm(ctx context.Context, p PaymentDocument, forms []PaymentForm) (int64, error) {
	amount, valid := new(big.Rat).SetString(p.Amount)
	if s == nil || s.DB == nil || p.ID == uuid.Nil || p.CompanyID <= 0 || p.ProviderID <= 0 || p.OpeningID <= 0 || p.CurrencyID <= 0 || !valid || amount.Sign() <= 0 || p.Date == "" || len(forms) == 0 {
		return 0, ErrInvalidPaymentRequest
	}
	tx := s.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	defer func() { _ = tx.Rollback() }()
	var id int64
	var err error
	query := tx.Raw(`INSERT INTO gf_pago_cabecera(empresa_id,uid,tercero_id,apertura_id,moneda_id,importe,fecha) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id`, p.CompanyID, p.ID, p.ProviderID, p.OpeningID, p.CurrencyID, p.Amount, p.Date)
	row := query.Row()
	if query.Error != nil {
		return 0, query.Error
	}
	if err = row.Scan(&id); err != nil {
		return 0, err
	}
	for _, form := range forms {
		value, ok := new(big.Rat).SetString(form.Amount)
		if form.MethodID <= 0 || !ok || value.Sign() <= 0 {
			return 0, ErrInvalidPaymentRequest
		}
		if result := tx.Exec(`INSERT INTO gf_pago_forma(empresa_id,cabecera_id,medio_id,importe,referencia) VALUES($1,$2,$3,$4,$5)`, p.CompanyID, id, form.MethodID, form.Amount, strings.TrimSpace(form.Reference)); result.Error != nil {
			return 0, result.Error
		}
	}
	var balanced bool
	query = tx.Raw(`SELECT COALESCE(SUM(importe),0) = $1::numeric FROM gf_pago_forma WHERE empresa_id=$2 AND cabecera_id=$3`, p.Amount, p.CompanyID, id)
	row = query.Row()
	err = query.Error
	if err == nil {
		err = row.Scan(&balanced)
	}
	if err != nil {
		return 0, err
	}
	if !balanced {
		return 0, ErrInvalidPaymentRequest
	}
	if result := tx.Exec(`SELECT gf_confirmar($1,$2,$3)`, p.CompanyID, "gf_pago", id); result.Error != nil {
		return 0, result.Error
	}
	if result := tx.Commit(); result.Error != nil {
		return 0, result.Error
	}
	return id, nil
}

func (s *PaymentStore) Create(ctx context.Context, p PaymentDocument) (int64, error) {
	amount, valid := new(big.Rat).SetString(p.Amount)
	if s == nil || s.DB == nil || p.ID == uuid.Nil || p.CompanyID <= 0 || p.ProviderID <= 0 || p.OpeningID <= 0 || p.CurrencyID <= 0 || !valid || amount.Sign() <= 0 || p.Date == "" {
		return 0, ErrInvalidPaymentRequest
	}
	var id int64
	query := s.DB.WithContext(ctx).Raw(`INSERT INTO gf_pago_cabecera(empresa_id,uid,tercero_id,apertura_id,moneda_id,importe,fecha) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id`, p.CompanyID, p.ID, p.ProviderID, p.OpeningID, p.CurrencyID, p.Amount, p.Date)
	row := query.Row()
	err := query.Error
	if err == nil {
		err = row.Scan(&id)
	}
	return id, err
}

func (s *PaymentStore) AddForm(ctx context.Context, companyID, headerID, methodID int64, amount, reference string) error {
	value, valid := new(big.Rat).SetString(amount)
	if s == nil || s.DB == nil || companyID <= 0 || headerID <= 0 || methodID <= 0 || !valid || value.Sign() <= 0 {
		return ErrInvalidPaymentRequest
	}
	return s.DB.WithContext(ctx).Exec(`INSERT INTO gf_pago_forma(empresa_id,cabecera_id,medio_id,importe,referencia) VALUES($1,$2,$3,$4,$5)`, companyID, headerID, methodID, amount, strings.TrimSpace(reference)).Error
}

func (s *PaymentStore) Confirm(ctx context.Context, companyID, headerID int64) error {
	if s == nil || s.DB == nil || companyID <= 0 || headerID <= 0 {
		return ErrInvalidPaymentRequest
	}
	return s.DB.WithContext(ctx).Exec(`SELECT gf_confirmar($1,$2,$3)`, companyID, "gf_pago", headerID).Error
}
