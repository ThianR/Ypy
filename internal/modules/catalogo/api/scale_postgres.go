package api

import (
	"context"
	"errors"
	"gorm.io/gorm"
)

// LoadV4ScaleRules lee la configuración versionada de balanza de una empresa.
func LoadV4ScaleRules(ctx context.Context, db *gorm.DB, empresaID int64) ([]V4ScaleRule, error) {
	if db == nil || empresaID <= 0 {
		return nil, errors.New("invalid scale rule store or company")
	}
	var rows []V4ScaleRule
	result := db.WithContext(ctx).Raw(`SELECT prefijo AS prefix,longitud AS length,inicio_producto AS product_start,largo_producto AS product_length,inicio_valor AS value_start,largo_valor AS value_length,decimales,contenido AS content FROM erp_v4.gp_regla_balanza WHERE empresa_id=$1 ORDER BY id`, empresaID).Scan(&rows)
	if result.Error != nil {
		return nil, result.Error
	}
	rules := make([]V4ScaleRule, 0, len(rows))
	for _, rule := range rows {
		if rule.Valid() {
			rules = append(rules, rule)
		}
	}
	return rules, nil
}
