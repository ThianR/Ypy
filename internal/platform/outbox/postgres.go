package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strconv"
)

// PostgresStore proporciona la parte durable del contrato de salida. La
// publicación es externa: el evento se marca solo después de ser aceptado,
// por lo que una interrupción lo deja disponible para reintento.
type PostgresStore struct{ DB *gorm.DB }

// ProcessPending publica eventos durables y confirma cada evento solo después
// de una entrega exitosa. Un fallo temporal detiene el lote y deja pendientes
// el evento fallido y los siguientes para la próxima ejecución.
func (s PostgresStore) ProcessPending(ctx context.Context, limit int, publisher Publisher) error {
	events, err := s.Pending(ctx, limit)
	if err != nil {
		return err
	}
	for i := range events {
		if err := publisher.Publish(events[i]); err != nil {
			_ = s.MarkFailed(ctx, events[i].ID, err)
			return err
		}
		if err := s.MarkPublished(ctx, events[i].ID); err != nil {
			return err
		}
	}
	return nil
}

// MarkFailed registra el último error de entrega sin confirmar el evento.
// La fila permanece disponible para reintento y su contador queda persistido.
func (s PostgresStore) MarkFailed(ctx context.Context, id string, publishErr error) error {
	if s.DB == nil || id == "" || publishErr == nil {
		return errors.New("invalid outbox event failure")
	}
	eventID, err := strconv.ParseInt(id, 10, 64)
	if err != nil || eventID <= 0 {
		return errors.New("invalid outbox event failure")
	}
	return s.DB.WithContext(ctx).Exec(`UPDATE erp_v4.gs_evento_salida SET reintentos=reintentos+1, publicado_error=$2 WHERE id=$1 AND publicado_en IS NULL`, eventID, fmt.Sprint(publishErr)).Error
}

func (s PostgresStore) Pending(ctx context.Context, limit int) ([]Event, error) {
	if s.DB == nil || limit <= 0 {
		return nil, errors.New("invalid outbox store or limit")
	}
	var rows []struct {
		ID      int64
		Topic   string `gorm:"column:tema"`
		Payload string `gorm:"column:contenido"`
	}
	result := s.DB.WithContext(ctx).Raw(`SELECT id, tema, contenido::text FROM erp_v4.gs_evento_salida WHERE publicado_en IS NULL ORDER BY id LIMIT $1`, limit).Scan(&rows)
	if result.Error != nil {
		return nil, result.Error
	}
	events := make([]Event, 0, len(rows))
	for _, row := range rows {
		event := Event{ID: strconv.FormatInt(row.ID, 10), Topic: row.Topic, Payload: row.Payload}
		if event.ID == "" || event.Topic == "" {
			return nil, errors.New("invalid outbox event")
		}
		if err := ValidatePayload(event.Payload); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

func (s PostgresStore) MarkPublished(ctx context.Context, id string) error {
	if s.DB == nil || id == "" {
		return errors.New("invalid outbox event")
	}
	eventID, err := strconv.ParseInt(id, 10, 64)
	if err != nil || eventID <= 0 {
		return errors.New("invalid outbox event")
	}
	return s.DB.WithContext(ctx).Exec(`UPDATE erp_v4.gs_evento_salida SET publicado_en=now() WHERE id=$1 AND publicado_en IS NULL`, eventID).Error
}

func ValidatePayload(payload string) error {
	if !json.Valid([]byte(payload)) {
		return errors.New("invalid outbox payload")
	}
	return nil
}
