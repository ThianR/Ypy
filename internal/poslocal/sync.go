package poslocal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"
)

var ErrSyncUnavailable = errors.New("central sync unavailable")

type syncAck struct {
	OperationID string `json:"operation_id"`
	ContentHash string `json:"content_hash"`
	Status      string `json:"status"`
}

// SyncPending publica operaciones locales durables en orden de creación.
// El identificador de operación se conserva como clave de idempotencia central en los reintentos.
func SyncPending(ctx context.Context, db *gorm.DB, baseURL, companyID, terminalID string, client *http.Client) (int, error) {
	if db == nil || strings.TrimSpace(baseURL) == "" || companyID == "" || terminalID == "" {
		return 0, ErrInvalidOperation
	}
	if client == nil {
		client = http.DefaultClient
	}
	type pending struct{ id, number, total, currency, created string }
	var stored []struct{ ID, Number, Total, Currency, State, Created string }
	if result := db.WithContext(ctx).Raw(`SELECT operation_id AS id,number,total,currency,state,created_at AS created FROM pos_operations WHERE state IN ('PENDING','SENDING') ORDER BY created_at,operation_id`).Scan(&stored); result.Error != nil {
		return 0, result.Error
	}
	pendingRows := make([]pending, 0, len(stored))
	for _, item := range stored {
		pendingRows = append(pendingRows, pending{id: item.ID, number: item.Number, total: item.Total, currency: item.Currency, created: item.Created})
	}
	sent := 0
	var err error
	for _, item := range pendingRows {
		id, number, total, currency, created := item.id, item.number, item.total, item.currency, item.created
		payload, _ := json.Marshal(map[string]string{"operation_id": id, "number": number, "total": total, "currency": currency, "created_at": created})
		digest := sha256.Sum256(payload)
		expectedHash := hex.EncodeToString(digest[:])
		result := db.WithContext(ctx).Exec(`UPDATE pos_operations SET state='SENDING' WHERE operation_id=? AND state IN ('PENDING','SENDING')`, id)
		err = result.Error
		if err != nil {
			return sent, err
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/api/v1/pos/operations", strings.NewReader(fmt.Sprintf(`{"id":%q,"empresa_id":%q,"terminal_id":%q,"payload":%q}`, id, companyID, terminalID, string(payload))))
		if err != nil {
			return sent, err
		}
		request.Header.Set("Content-Type", "application/json")
		var response *http.Response
		for attempt := 0; attempt < 3; attempt++ {
			if attempt > 0 {
				delay := time.Duration(100*(1<<(attempt-1))) * time.Millisecond
				timer := time.NewTimer(delay)
				select {
				case <-ctx.Done():
					timer.Stop()
					return sent, ctx.Err()
				case <-timer.C:
				}
			}
			if attempt > 0 && request.GetBody != nil {
				request.Body, err = request.GetBody()
				if err != nil {
					break
				}
			}
			response, err = client.Do(request)
			if err == nil && response.StatusCode >= 200 && response.StatusCode < 300 {
				break
			}
			retryable := err != nil || (response != nil && (response.StatusCode == 408 || response.StatusCode == 429 || response.StatusCode >= 500))
			if response != nil {
				_ = response.Body.Close()
			}
			if !retryable {
				break
			}
		}
		if err != nil || response == nil || response.StatusCode < 200 || response.StatusCode >= 300 {
			if response != nil {
				_ = response.Body.Close()
			}
			_ = db.WithContext(ctx).Exec(`UPDATE pos_operations SET state='PENDING' WHERE operation_id=?`, id).Error
			if err != nil {
				return sent, err
			}
			return sent, ErrSyncUnavailable
		}
		var ack syncAck
		err = json.NewDecoder(response.Body).Decode(&ack)
		_ = response.Body.Close()
		if err != nil || ack.OperationID != id || ack.Status != "ACCEPTED" || ack.ContentHash != expectedHash {
			_ = db.WithContext(ctx).Exec(`UPDATE pos_operations SET state='CONFLICT' WHERE operation_id=?`, id).Error
			return sent, ErrSyncUnavailable
		}
		if result := db.WithContext(ctx).Exec(`UPDATE pos_operations SET state='APPLIED' WHERE operation_id=?`, id); result.Error != nil {
			err = result.Error
			return sent, err
		}
		sent++
	}
	return sent, nil
}
