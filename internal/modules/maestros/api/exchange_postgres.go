package api

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"regexp"
	"time"
)

var ErrInvalidQuoteScope = errors.New("invalid quote scope")

type QuoteStore struct{ DB *gorm.DB }

var quoteRate = regexp.MustCompile(`^(0|[1-9][0-9]*)(\.[0-9]+)?$`)

func validQuoteKind(kind string) bool {
	return kind == "COMPRA" || kind == "VENTA" || kind == "CONTABLE"
}

// SaveQuote appends a dated quote. Historical values are never updated.
func (s *QuoteStore) SaveQuote(ctx context.Context, companyID, fromID, toID int64, kind, rate string, effectiveAt time.Time) error {
	if s == nil || s.DB == nil || companyID <= 0 || fromID <= 0 || toID <= 0 || fromID == toID || !validQuoteKind(kind) || !quoteRate.MatchString(rate) || rate == "0" || effectiveAt.IsZero() {
		return ErrInvalidQuoteScope
	}
	return s.DB.WithContext(ctx).Exec(`INSERT INTO gs_cotizacion_moneda(empresa_id,moneda_origen_id,moneda_destino_id,fecha,tipo,valor) VALUES($1,$2,$3,$4,$5,$6)`, companyID, fromID, toID, effectiveAt, kind, rate).Error
}

// LoadQuote devuelve la cotización más reciente vigente en asOf. El AsOf
// devuelto es la marca de tiempo de la base y permite conservar el snapshot.
func (s *QuoteStore) LoadQuote(ctx context.Context, companyID, fromID, toID int64, kind string, asOf time.Time) (Quote, error) {
	if s == nil || s.DB == nil || companyID <= 0 || fromID <= 0 || toID <= 0 || fromID == toID || !validQuoteKind(kind) || asOf.IsZero() {
		return Quote{}, ErrInvalidQuoteScope
	}
	var q Quote
	query := s.DB.WithContext(ctx).Raw(`SELECT o.codigo AS from,d.codigo AS to,c.valor AS rate,c.fecha AS as_of
FROM gs_cotizacion_moneda c
JOIN gs_moneda o ON o.id=c.moneda_origen_id
JOIN gs_moneda d ON d.id=c.moneda_destino_id
WHERE c.empresa_id=$1 AND c.moneda_origen_id=$2 AND c.moneda_destino_id=$3 AND c.tipo=$4 AND c.fecha<=$5
	ORDER BY c.fecha DESC LIMIT 1`, companyID, fromID, toID, kind, asOf).Scan(&q)
	if query.Error != nil {
		return Quote{}, query.Error
	}
	if query.RowsAffected == 0 {
		return Quote{}, gorm.ErrRecordNotFound
	}
	return q, nil
}
