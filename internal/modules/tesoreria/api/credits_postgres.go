package api

import (
	"context"
	"errors"
	"math/big"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrInvalidCreditApplication = errors.New("invalid credit application")

type CreditStore struct{ DB *gorm.DB }

func (s *CreditStore) Apply(ctx context.Context, companyID, creditID, installmentID int64, amount string, id uuid.UUID) (int64, error) {
	value, valid := new(big.Rat).SetString(amount)
	if s == nil || s.DB == nil || companyID <= 0 || creditID <= 0 || installmentID <= 0 || id == uuid.Nil || !valid || value.Sign() <= 0 {
		return 0, ErrInvalidCreditApplication
	}
	var applicationID int64
	query := s.DB.WithContext(ctx).Raw(`SELECT erp_v4.gf_usar_credito($1,$2,$3,$4,$5,$6)`, companyID, "gv", creditID, installmentID, amount, id)
	row := query.Row()
	if query.Error != nil {
		return 0, query.Error
	}
	if err := row.Scan(&applicationID); err != nil {
		return 0, err
	}
	return applicationID, nil
}
