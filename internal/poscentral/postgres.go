package poscentral

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ypy-erp/ypy/internal/platform/events"
	"gorm.io/gorm"
)

// PostgresStore persiste la parte de entrada del contrato POS en V4.
// La aplicación comercial permanece en una transacción separada después de aceptar la entrada.
type PostgresStore struct{ DB *gorm.DB }

func posRow(db *gorm.DB, query string, args ...any) (*sql.Row, error) {
	result := db.Raw(query, args...)
	return result.Row(), result.Error
}

func (s PostgresStore) EnqueueAccepted(ctx context.Context, op Operation, occurred time.Time) error {
	if s.DB == nil || Validate(op) != nil {
		return ErrInvalidOperation
	}
	company, err := strconv.ParseInt(op.EmpresaID, 10, 64)
	if err != nil || company <= 0 {
		return ErrInvalidOperation
	}
	eventID := uuid.New()
	envelope := events.Envelope{EventID: eventID.String(), EventType: "pos.operation.accepted", EmpresaID: op.EmpresaID, Version: 1, Origin: op.TerminalID, CorrelationID: op.ID, OccurredAt: occurred.UTC().Format(time.RFC3339), Payload: json.RawMessage(op.Payload)}
	if err := envelope.Validate(); err != nil {
		return err
	}
	content, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	return s.DB.WithContext(ctx).Exec(`INSERT INTO erp_v4.gs_evento_salida(empresa_id,uid,tema,contenido) VALUES($1,$2,$3,$4::jsonb) ON CONFLICT (uid) DO NOTHING`, company, eventID, envelope.EventType, content).Error
}

func (s PostgresStore) SaveCursor(ctx context.Context, companyID, terminalID int64, cursor int64) error {
	if s.DB == nil || companyID <= 0 || terminalID <= 0 || cursor < 0 {
		return errors.New("invalid sync cursor")
	}
	return s.DB.WithContext(ctx).Exec(`INSERT INTO erp_v4.gp_sync_cursor(empresa_id,terminal_id,cursor) VALUES($1,$2,$3) ON CONFLICT(empresa_id,terminal_id) DO UPDATE SET cursor=GREATEST(erp_v4.gp_sync_cursor.cursor,EXCLUDED.cursor), actualizado_en=now()`, companyID, terminalID, cursor).Error
}

func (s PostgresStore) Cursor(ctx context.Context, companyID, terminalID int64) (int64, error) {
	if s.DB == nil || companyID <= 0 || terminalID <= 0 {
		return 0, errors.New("nil postgres database")
	}
	var cursor int64
	row, err := posRow(s.DB.WithContext(ctx), `SELECT cursor FROM erp_v4.gp_sync_cursor WHERE empresa_id=$1 AND terminal_id=$2`, companyID, terminalID)
	if err == nil {
		err = row.Scan(&cursor)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return cursor, err
}

// IntegrateSale ejecuta la función de venta de V4, que administra numeración,
// confirmación, movimiento de stock y estado de entrada en una transacción.
func (s PostgresStore) IntegrateSale(ctx context.Context, op Operation) (int64, error) {
	if s.DB == nil {
		return 0, errors.New("nil postgres database")
	}
	var data struct {
		Sequence int64  `json:"sequence"`
		Occurred string `json:"occurred_at"`
	}
	if err := json.Unmarshal([]byte(op.Payload), &data); err != nil || data.Sequence <= 0 || data.Occurred == "" {
		return 0, ErrInvalidOperation
	}
	if err := validateSalePayload(op.Payload); err != nil {
		return 0, err
	}
	company, err := strconv.ParseInt(op.EmpresaID, 10, 64)
	if err != nil || company <= 0 {
		return 0, fmt.Errorf("empresa_id: %w", err)
	}
	terminal, err := strconv.ParseInt(op.TerminalID, 10, 64)
	if err != nil || terminal <= 0 {
		return 0, fmt.Errorf("terminal_id: %w", err)
	}
	uid, err := uuid.Parse(op.ID)
	if err != nil {
		return 0, fmt.Errorf("id: %w", err)
	}
	when, err := time.Parse(time.RFC3339, data.Occurred)
	if err != nil {
		return 0, fmt.Errorf("occurred_at: %w", err)
	}
	var saleID sql.NullInt64
	row, err := posRow(s.DB.WithContext(ctx),
		`SELECT erp_v4.gp_integrar_venta($1,$2,$3,$4,$5,$6::jsonb)`,
		company, terminal, data.Sequence, uid, when, op.Payload)
	if err == nil {
		err = row.Scan(&saleID)
	}
	if err != nil {
		return 0, err
	}
	if !saleID.Valid {
		return 0, errors.New("sale integration returned no id")
	}
	return saleID.Int64, nil
}

// IntegrateSaleWithReceipt integra venta y cobro en una única transacción V4.
func (s PostgresStore) IntegrateSaleWithReceipt(ctx context.Context, op Operation) (int64, error) {
	if s.DB == nil {
		return 0, errors.New("nil postgres database")
	}
	var data struct {
		Sequence int64             `json:"sequence"`
		Occurred string            `json:"occurred_at"`
		Forms    []json.RawMessage `json:"formas"`
	}
	if err := json.Unmarshal([]byte(op.Payload), &data); err != nil || data.Sequence <= 0 || data.Occurred == "" || len(data.Forms) == 0 {
		return 0, ErrInvalidOperation
	}
	if err := validateSalePayload(op.Payload); err != nil {
		return 0, err
	}
	company, err := strconv.ParseInt(op.EmpresaID, 10, 64)
	if err != nil || company <= 0 {
		return 0, fmt.Errorf("empresa_id: %w", err)
	}
	terminal, err := strconv.ParseInt(op.TerminalID, 10, 64)
	if err != nil || terminal <= 0 {
		return 0, fmt.Errorf("terminal_id: %w", err)
	}
	uid, err := uuid.Parse(op.ID)
	if err != nil {
		return 0, fmt.Errorf("id: %w", err)
	}
	when, err := time.Parse(time.RFC3339, data.Occurred)
	if err != nil {
		return 0, fmt.Errorf("occurred_at: %w", err)
	}
	var saleID sql.NullInt64
	row, err := posRow(s.DB.WithContext(ctx), `SELECT erp_v4.gp_integrar_venta_cobro($1,$2,$3,$4,$5,$6::jsonb)`, company, terminal, data.Sequence, uid, when, op.Payload)
	if err == nil {
		err = row.Scan(&saleID)
	}
	if err != nil {
		return 0, err
	}
	if !saleID.Valid {
		return 0, errors.New("sale and receipt integration returned no id")
	}
	return saleID.Int64, nil
}

// validateSalePayload evita que las conversiones SQL de gp_integrar_venta sean
// la primera frontera de validación. Solo verifica los campos que la función
// V4 desreferencia o recorre estructuralmente.
func validateSalePayload(payload string) error {
	var data struct {
		PermisoID int64             `json:"permiso_id"`
		UsuarioID int64             `json:"usuario_id"`
		Talonario int64             `json:"talonario_id"`
		RangoID   int64             `json:"rango_id"`
		Numero    int64             `json:"numero"`
		TipoID    int64             `json:"tipo_id"`
		MonedaID  int64             `json:"moneda_id"`
		Lineas    []json.RawMessage `json:"lineas"`
	}
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return ErrInvalidOperation
	}
	if data.PermisoID <= 0 || data.UsuarioID <= 0 || data.Talonario <= 0 || data.RangoID <= 0 || data.Numero <= 0 || data.TipoID <= 0 || data.MonedaID <= 0 || len(data.Lineas) == 0 {
		return ErrInvalidOperation
	}
	return nil
}

func (s PostgresStore) Apply(op Operation) (Operation, error) {
	return s.ApplyContext(context.Background(), op)
}

func (s PostgresStore) ApplyContext(ctx context.Context, op Operation) (Operation, error) {
	if s.DB == nil {
		return Operation{}, errors.New("nil postgres database")
	}
	company, err := strconv.ParseInt(op.EmpresaID, 10, 64)
	if err != nil || company <= 0 {
		return Operation{}, fmt.Errorf("empresa_id: %w", err)
	}
	terminal, err := strconv.ParseInt(op.TerminalID, 10, 64)
	if err != nil || terminal <= 0 {
		return Operation{}, fmt.Errorf("terminal_id: %w", err)
	}
	uid, err := uuid.Parse(op.ID)
	if err != nil {
		return Operation{}, fmt.Errorf("id: %w", err)
	}
	var payload json.RawMessage
	trimmedPayload := strings.TrimSpace(op.Payload)
	if !json.Valid([]byte(trimmedPayload)) || len(trimmedPayload) == 0 || trimmedPayload[0] != '{' {
		return Operation{}, ErrInvalidOperation
	}
	payload = json.RawMessage(trimmedPayload)
	op.Payload = trimmedPayload
	h := hash(op.Payload)
	sequence, ok := opSequence(op.Payload)
	if !ok {
		return Operation{}, ErrInvalidOperation
	}
	result := s.DB.WithContext(ctx).Exec(`
INSERT INTO erp_v4.gp_evento_entrada
 (empresa_id, uid, terminal_id, secuencia, ocurrido_en, contenido, contenido_hash)
VALUES ($1,$2,$3,$4,$5,$6::json,$7)
	ON CONFLICT (uid) DO NOTHING`, company, uid, terminal, sequence, time.Now().UTC(), payload, h)
	if result.Error != nil {
		return Operation{}, result.Error
	}
	var storedCompany, storedTerminal, storedSequence int64
	var storedHash string
	row, err := posRow(s.DB.WithContext(ctx), `SELECT empresa_id, terminal_id, secuencia, contenido_hash FROM erp_v4.gp_evento_entrada WHERE uid=$1`, uid)
	if err == nil {
		err = row.Scan(&storedCompany, &storedTerminal, &storedSequence, &storedHash)
	}
	if err != nil {
		return Operation{}, err
	}
	if storedCompany != company || storedTerminal != terminal || storedSequence != sequence || storedHash != h {
		return Operation{}, ErrConflict
	}
	return op, nil
}

func opSequence(payload string) (int64, bool) {
	var v struct {
		Sequence int64 `json:"sequence"`
	}
	if json.Unmarshal([]byte(payload), &v) != nil || v.Sequence <= 0 {
		return 0, false
	}
	return v.Sequence, true
}
