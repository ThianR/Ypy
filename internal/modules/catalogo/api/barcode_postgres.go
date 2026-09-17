package api

import (
	"context"
	"errors"
	"regexp"
	"time"

	"gorm.io/gorm"
)

var ErrInvalidBarcodeScope = errors.New("invalid barcode scope")
var ErrBarcodeNotFound = errors.New("barcode not found")

type BarcodeMatch struct {
	CompanyID        int64  `json:"empresa_id"`
	ArticleID        int64  `json:"articulo_id"`
	PresentationID   int64  `json:"presentacion_id"`
	ArticleCode      string `json:"articulo_codigo"`
	Description      string `json:"descripcion"`
	PresentationCode string `json:"presentacion_codigo"`
	FactorBase       string `json:"factor_base"`
}
type CatalogSearchItem struct {
	CompanyID        int64  `json:"empresa_id"`
	ArticleID        int64  `json:"articulo_id"`
	PresentationID   int64  `json:"presentacion_id"`
	Barcode          string `json:"barcode"`
	ArticleCode      string `json:"articulo_codigo"`
	Description      string `json:"descripcion"`
	PresentationCode string `json:"presentacion_codigo"`
	Price            string `json:"precio"`
}
type BarcodeStore struct{ DB *gorm.DB }

var positiveQuantity = regexp.MustCompile(`^(0|[1-9][0-9]*)(\.[0-9]+)?$`)

func (s *BarcodeStore) Price(ctx context.Context, companyID, listID, presentationID int64, quantity string, at time.Time) (string, error) {
	if s == nil || s.DB == nil || companyID <= 0 || listID <= 0 || presentationID <= 0 || !positiveQuantity.MatchString(quantity) || quantity == "0" || at.IsZero() {
		return "", ErrInvalidBarcodeScope
	}
	var price string
	result := s.DB.WithContext(ctx).Raw(`SELECT precio::text FROM erp_v4.gv_precio WHERE empresa_id=$1 AND lista_id=$2 AND presentacion_id=$3 AND desde<= $4 AND (hasta IS NULL OR $4<hasta) AND minimo<= $5 ORDER BY minimo DESC,desde DESC LIMIT 1`, companyID, listID, presentationID, at, quantity).Scan(&price)
	if result.Error != nil {
		return "", result.Error
	}
	if result.RowsAffected == 0 {
		return "", ErrBarcodeNotFound
	}
	return price, nil
}

func (s *BarcodeStore) Resolve(ctx context.Context, companyID int64, barcode string) (BarcodeMatch, error) {
	if s == nil || s.DB == nil || companyID <= 0 || barcode == "" {
		return BarcodeMatch{}, ErrInvalidBarcodeScope
	}
	var m BarcodeMatch
	result := s.DB.WithContext(ctx).Raw(`SELECT b.empresa_id AS company_id,a.id AS article_id,p.id AS presentation_id,a.codigo AS article_code,a.descripcion,p.codigo AS presentation_code,p.factor_base
FROM erp_v4.gi_codigo_barra b JOIN erp_v4.gi_presentacion p ON p.empresa_id=b.empresa_id AND p.id=b.presentacion_id
JOIN erp_v4.gi_articulo a ON a.empresa_id=p.empresa_id AND a.id=p.articulo_id
	WHERE b.empresa_id=$1 AND b.codigo=$2 AND p.activa AND a.activo AND a.vendible`, companyID, barcode).Scan(&m)
	if result.Error != nil {
		return BarcodeMatch{}, result.Error
	}
	if result.RowsAffected == 0 {
		return BarcodeMatch{}, ErrBarcodeNotFound
	}
	return m, nil
}

func (s *BarcodeStore) Search(ctx context.Context, companyID, listID int64, query string, at time.Time, limit int) ([]CatalogSearchItem, error) {
	if s == nil || s.DB == nil || companyID <= 0 || listID <= 0 || query == "" || at.IsZero() || limit <= 0 || limit > 100 {
		return nil, ErrInvalidBarcodeScope
	}
	var items []CatalogSearchItem
	result := s.DB.WithContext(ctx).Raw(`SELECT b.codigo AS barcode,a.id AS article_id,p.id AS presentation_id,a.codigo AS article_code,a.descripcion AS description,p.codigo AS presentation_code,coalesce(pr.precio::text,'') AS price
FROM erp_v4.gi_codigo_barra b JOIN erp_v4.gi_presentacion p ON p.empresa_id=b.empresa_id AND p.id=b.presentacion_id
JOIN erp_v4.gi_articulo a ON a.empresa_id=p.empresa_id AND a.id=p.articulo_id
LEFT JOIN LATERAL (SELECT precio FROM erp_v4.gv_precio WHERE empresa_id=$1 AND lista_id=$2 AND presentacion_id=p.id AND desde<= $4 AND (hasta IS NULL OR $4<hasta) AND minimo<=1 ORDER BY desde DESC,minimo DESC LIMIT 1) pr ON true
WHERE b.empresa_id=$1 AND p.activa AND a.activo AND a.vendible AND (b.codigo ILIKE '%'||$3||'%' OR a.codigo ILIKE '%'||$3||'%' OR a.descripcion ILIKE '%'||$3||'%')
	ORDER BY a.descripcion,b.codigo LIMIT $5`, companyID, listID, query, at, limit).Scan(&items)
	if result.Error != nil {
		return nil, result.Error
	}
	for i := range items {
		item := &items[i]
		item.CompanyID = companyID
	}
	return items, nil
}
