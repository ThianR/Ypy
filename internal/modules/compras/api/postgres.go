package api

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

var ErrInvalidReceiptScope = errors.New("invalid receipt scope")
var ErrReceiptHeaderConflict = errors.New("receipt header conflict")
var ErrReceiptDetailConflict = errors.New("receipt detail conflict")

type PostgresStore struct{ DB *gorm.DB }

func purchaseRow(db *gorm.DB, query string, args ...any) (*sql.Row, error) {
	result := db.Raw(query, args...)
	return result.Row(), result.Error
}

type ReceiptHeader struct {
	ID                                                                      uuid.UUID
	CompanyID, BranchID, UserID, TypeID, SupplierID, CurrencyID, TerminalID int64
	Total, Snapshot                                                         string
	At                                                                      time.Time
}
type ReceiptDetail struct {
	CompanyID, HeaderID, Line, ArticleID, PresentationID int64
	Description, Quantity, FactorBase, Price             string
}

// CreateHeader inserta una recepción V4 en estado borrador. Los reintentos UUID devuelven el
// existing header, preserving idempotency without applying stock twice.
func (s *PostgresStore) CreateHeader(ctx context.Context, h ReceiptHeader) (int64, error) {
	if s == nil || s.DB == nil || h.ID == uuid.Nil || h.CompanyID <= 0 || h.BranchID <= 0 || h.UserID <= 0 || h.TypeID <= 0 || h.SupplierID <= 0 || h.CurrencyID <= 0 || h.TerminalID <= 0 || h.Total == "" || h.At.IsZero() {
		return 0, ErrInvalidReceiptScope
	}
	var id, company, branch, user, typ, terminal, supplier, currency int64
	var total, snapshot string
	var at time.Time
	row, err := purchaseRow(s.DB.WithContext(ctx), `SELECT id,empresa_id,sucursal_id,usuario_id,tipo_id,terminal_id,tercero_id,moneda_id,fecha_operacion,total,tercero_snapshot::text FROM gc_recepcion_cabecera WHERE uid=$1`, h.ID)
	if err == nil {
		err = row.Scan(&id, &company, &branch, &user, &typ, &terminal, &supplier, &currency, &at, &total, &snapshot)
	}
	if err == nil {
		if company != h.CompanyID || branch != h.BranchID || user != h.UserID || typ != h.TypeID || terminal != h.TerminalID || supplier != h.SupplierID || currency != h.CurrencyID || total != h.Total || snapshot != h.Snapshot || !at.Equal(h.At) {
			return 0, ErrReceiptHeaderConflict
		}
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	row, err = purchaseRow(s.DB.WithContext(ctx), `INSERT INTO gc_recepcion_cabecera(empresa_id,uid,sucursal_id,usuario_id,tipo_id,terminal_id,tercero_id,moneda_id,fecha_operacion,total,tercero_snapshot) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id`, h.CompanyID, h.ID, h.BranchID, h.UserID, h.TypeID, h.TerminalID, h.SupplierID, h.CurrencyID, h.At, h.Total, h.Snapshot)
	if err == nil {
		err = row.Scan(&id)
	}
	return id, err
}

func (s *PostgresStore) AddDetail(ctx context.Context, d ReceiptDetail) (int64, error) {
	if s == nil || s.DB == nil || d.CompanyID <= 0 || d.HeaderID <= 0 || d.Line <= 0 || d.ArticleID <= 0 || d.PresentationID <= 0 || d.Description == "" || d.Quantity == "" || d.FactorBase == "" || d.Price == "" {
		return 0, ErrInvalidReceiptScope
	}
	var existingID, existingArticle, existingPresentation int64
	var existingQuantity, existingFactor, existingPrice, existingDescription string
	row, err := purchaseRow(s.DB.WithContext(ctx), `SELECT id,articulo_id,presentacion_id,descripcion,cantidad::text,factor_base::text,precio::text FROM gc_recepcion_detalle WHERE empresa_id=$1 AND cabecera_id=$2 AND renglon=$3`, d.CompanyID, d.HeaderID, d.Line)
	if err == nil {
		err = row.Scan(&existingID, &existingArticle, &existingPresentation, &existingDescription, &existingQuantity, &existingFactor, &existingPrice)
	}
	if err == nil {
		if existingArticle != d.ArticleID || existingPresentation != d.PresentationID || existingDescription != d.Description || existingQuantity != d.Quantity || existingFactor != d.FactorBase || existingPrice != d.Price {
			return 0, ErrReceiptDetailConflict
		}
		return existingID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	var id int64
	row, err = purchaseRow(s.DB.WithContext(ctx), `INSERT INTO gc_recepcion_detalle(empresa_id,cabecera_id,renglon,articulo_id,presentacion_id,descripcion,cantidad,factor_base,precio) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`, d.CompanyID, d.HeaderID, d.Line, d.ArticleID, d.PresentationID, d.Description, d.Quantity, d.FactorBase, d.Price)
	if err == nil {
		err = row.Scan(&id)
	}
	return id, err
}

func (s *PostgresStore) Confirm(ctx context.Context, companyID, headerID int64) error {
	if s == nil || s.DB == nil || companyID <= 0 || headerID <= 0 {
		return ErrInvalidReceiptScope
	}
	return s.DB.WithContext(ctx).Exec(`SELECT gs_confirmar($1,$2,$3)`, companyID, "gc_recepcion", headerID).Error
}
