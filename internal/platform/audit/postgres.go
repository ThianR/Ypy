package audit

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostgresWriter struct{ DB *gorm.DB }

func (w PostgresWriter) Record(ctx context.Context, companyID, userID, recordID int64, table string, entry Entry) error {
	if w.DB == nil || table == "" || entry.OperationID == "" {
		return errors.New("invalid audit context")
	}
	operationID, err := uuid.Parse(entry.OperationID)
	if err != nil {
		return err
	}
	before, err := json.Marshal(entry.Before)
	if err != nil {
		return err
	}
	after, err := json.Marshal(entry.After)
	if err != nil {
		return err
	}
	return w.DB.WithContext(ctx).Exec(`INSERT INTO erp_v4.gs_auditoria (empresa_id,usuario_id,tabla,registro_id,accion,antes,despues,operation_id,motivo) VALUES ($1,$2,$3,$4,$5,$6::jsonb,$7::jsonb,$8,$9)`, companyID, userID, table, recordID, entry.Action, before, after, operationID, entry.Reason).Error
}
