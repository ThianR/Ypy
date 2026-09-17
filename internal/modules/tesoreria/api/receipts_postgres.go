package api

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"regexp"
)

var ErrInvalidReceiptRequest = errors.New("invalid receipt request")
var ErrReceiptConflict = errors.New("receipt UUID content conflict")
var receiptAmount = regexp.MustCompile(`^(0|[1-9][0-9]*)(\.[0-9]+)?$`)

type Receipt struct {
	ID                                           uuid.UUID
	CompanyID, CustomerID, OpeningID, CurrencyID int64
	Amount, Date                                 string
}
type ReceiptForm struct {
	CompanyID int64  `json:"empresa_id"`
	HeaderID  int64  `json:"cabecera_id"`
	MethodID  int64  `json:"medio_id"`
	Amount    string `json:"importe"`
	Reference string `json:"referencia"`
}
type ReceiptStore struct{ DB *gorm.DB }

func (s *ReceiptStore) Create(ctx context.Context, r Receipt) (int64, error) {
	if s == nil || s.DB == nil || r.ID == uuid.Nil || r.CompanyID <= 0 || r.CustomerID <= 0 || r.OpeningID <= 0 || r.CurrencyID <= 0 || r.Amount == "" || r.Date == "" {
		return 0, ErrInvalidReceiptRequest
	}
	var id, company, customer, opening, currency int64
	var amount, date string
	query := s.DB.WithContext(ctx).Raw(`SELECT id,empresa_id,tercero_id,apertura_id,moneda_id,importe::text,fecha::text FROM gf_recibo_cabecera WHERE uid=$1`, r.ID)
	row := query.Row()
	err := query.Error
	if err == nil {
		err = row.Scan(&id, &company, &customer, &opening, &currency, &amount, &date)
	}
	if err == nil {
		if company != r.CompanyID || customer != r.CustomerID || opening != r.OpeningID || currency != r.CurrencyID || amount != r.Amount || date != r.Date {
			return 0, ErrReceiptConflict
		}
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	query = s.DB.WithContext(ctx).Raw(`INSERT INTO gf_recibo_cabecera(empresa_id,uid,tercero_id,apertura_id,moneda_id,importe,fecha) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id`, r.CompanyID, r.ID, r.CustomerID, r.OpeningID, r.CurrencyID, r.Amount, r.Date)
	row = query.Row()
	err = query.Error
	if err == nil {
		err = row.Scan(&id)
	}
	return id, err
}

func (s *ReceiptStore) AddForm(ctx context.Context, f ReceiptForm) error {
	if s == nil || s.DB == nil || f.CompanyID <= 0 || f.HeaderID <= 0 || f.MethodID <= 0 || f.Amount == "" {
		return ErrInvalidReceiptRequest
	}
	return s.DB.WithContext(ctx).Exec(`INSERT INTO gf_recibo_forma(empresa_id,cabecera_id,medio_id,importe,referencia) VALUES($1,$2,$3,$4,$5)`, f.CompanyID, f.HeaderID, f.MethodID, f.Amount, f.Reference).Error
}

// Confirm confirma un recibo mediante la función transaccional de V4.
// La función central hace que la repetición sea idempotente.
func (s *ReceiptStore) Confirm(ctx context.Context, companyID, headerID int64) error {
	if s == nil || s.DB == nil || companyID <= 0 || headerID <= 0 {
		return ErrInvalidReceiptRequest
	}
	return s.DB.WithContext(ctx).Exec(`SELECT gf_confirmar($1,$2,$3)`, companyID, "gf_recibo", headerID).Error
}

// CreateAndConfirm crea un recibo, registra sus formas y lo confirma en una
// sola transacción. La base V4 conserva la autoridad sobre el estado final.
func (s *ReceiptStore) CreateAndConfirm(ctx context.Context, r Receipt, forms []ReceiptForm) (int64, error) {
	if s == nil || s.DB == nil || r.ID == uuid.Nil || r.CompanyID <= 0 || r.CustomerID <= 0 || r.OpeningID <= 0 || r.CurrencyID <= 0 || !receiptAmount.MatchString(r.Amount) || r.Date == "" || len(forms) == 0 {
		return 0, ErrInvalidReceiptRequest
	}
	tx := s.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		err := tx.Error
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	var id int64
	var company, customer, opening, currency int64
	var amount, date, state string
	query := tx.Raw(`SELECT id,empresa_id,tercero_id,apertura_id,moneda_id,importe::text,fecha::text,estado FROM gf_recibo_cabecera WHERE uid=$1 FOR UPDATE`, r.ID)
	row := query.Row()
	err := query.Error
	if err == nil {
		err = row.Scan(&id, &company, &customer, &opening, &currency, &amount, &date, &state)
	}
	if err == nil {
		if company != r.CompanyID || customer != r.CustomerID || opening != r.OpeningID || currency != r.CurrencyID || amount != r.Amount || date != r.Date {
			return 0, ErrReceiptConflict
		}
		if state == "CONFIRMADO" {
			if result := tx.Commit(); result.Error != nil {
				return 0, result.Error
			}
			return id, nil
		}
		return 0, ErrInvalidReceiptRequest
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	query = tx.Raw(`INSERT INTO gf_recibo_cabecera(empresa_id,uid,tercero_id,apertura_id,moneda_id,importe,fecha) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id`, r.CompanyID, r.ID, r.CustomerID, r.OpeningID, r.CurrencyID, r.Amount, r.Date)
	row = query.Row()
	err = query.Error
	if err == nil {
		err = row.Scan(&id)
	}
	if err != nil {
		return 0, err
	}
	for _, form := range forms {
		if form.CompanyID != r.CompanyID || (form.HeaderID != 0 && form.HeaderID != id) || form.MethodID <= 0 || !receiptAmount.MatchString(form.Amount) {
			return 0, ErrInvalidReceiptRequest
		}
		if result := tx.Exec(`INSERT INTO gf_recibo_forma(empresa_id,cabecera_id,medio_id,importe,referencia) VALUES($1,$2,$3,$4,$5)`, form.CompanyID, id, form.MethodID, form.Amount, form.Reference); result.Error != nil {
			err = result.Error
			return 0, err
		}
	}
	var balanced bool
	query = tx.Raw(`SELECT COALESCE(SUM(importe),0) = $1::numeric FROM gf_recibo_forma WHERE empresa_id=$2 AND cabecera_id=$3`, r.Amount, r.CompanyID, id)
	row = query.Row()
	err = query.Error
	if err == nil {
		err = row.Scan(&balanced)
	}
	if err != nil || !balanced {
		if err != nil {
			return 0, err
		}
		return 0, ErrInvalidReceiptRequest
	}
	if result := tx.Exec(`SELECT gf_confirmar($1,$2,$3)`, r.CompanyID, "gf_recibo", id); result.Error != nil {
		err = result.Error
		return 0, err
	}
	if result := tx.Commit(); result.Error != nil {
		return 0, result.Error
	}
	return id, nil
}
