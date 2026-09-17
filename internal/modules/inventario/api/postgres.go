package api

import (
	"context"
	"errors"
	"gorm.io/gorm"
)

var ErrInvalidScope = errors.New("invalid inventory scope")

type PostgresStore struct{ DB *gorm.DB }

// BalanceCurrent lee el saldo actual autorizado de la posición desde V4.
// No deriva el stock a partir de los totales históricos de movimientos.
func (s *PostgresStore) BalanceCurrent(ctx context.Context, companyID, itemID int64) (int64, error) {
	if s == nil || s.DB == nil || companyID <= 0 || itemID <= 0 {
		return 0, ErrInvalidScope
	}
	var balance int64
	result := s.DB.WithContext(ctx).Raw(`SELECT COALESCE(SUM(e.cantidad),0)::bigint
FROM gi_existencia e
JOIN gi_posicion p ON p.id=e.posicion_id AND p.empresa_id=e.empresa_id
WHERE e.empresa_id=$1 AND p.articulo_id=$2`, companyID, itemID).Scan(&balance)
	return balance, result.Error
}
