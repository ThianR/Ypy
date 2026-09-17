package poslocal

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func testAck(w http.ResponseWriter, r *http.Request, id string) {
	var request struct {
		Payload string `json:"payload"`
	}
	body, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(body, &request)
	digest := sha256.Sum256([]byte(request.Payload))
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"operation_id":"` + id + `","content_hash":"` + hex.EncodeToString(digest[:]) + `","status":"ACCEPTED"}`))
}

func TestSyncPendingPublishesAndMarksApplied(t *testing.T) {
	db, err := OpenStore(t.TempDir() + "/pos.db")
	if err != nil {
		t.Fatal(err)
	}
	defer CloseAgentDB(db)
	if err := SaveOperation(db, "op-sync", "1001", "10", "PYG", "PENDING", "2026-09-16T12:00:00Z"); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { testAck(w, r, "op-sync") }))
	defer server.Close()
	sent, err := SyncPending(context.Background(), db, server.URL, "4", "2", server.Client())
	if err != nil || sent != 1 {
		t.Fatalf("sent=%d err=%v", sent, err)
	}
	var state string
	row, queryErr := agentRow(db, `SELECT state FROM pos_operations WHERE operation_id='op-sync'`)
	if err := queryErr; err != nil || row.Scan(&state) != nil || state != "APPLIED" {
		t.Fatalf("state=%q err=%v", state, err)
	}
}

func TestSyncPendingRetriesTransientFailure(t *testing.T) {
	db, err := OpenStore(t.TempDir() + "/pos.db")
	if err != nil {
		t.Fatal(err)
	}
	defer CloseAgentDB(db)
	if err := SaveOperation(db, "op-retry", "1002", "10", "PYG", "PENDING", "2026-09-16T12:00:00Z"); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			http.Error(w, "temporary", http.StatusServiceUnavailable)
			return
		}
		testAck(w, r, "op-retry")
	}))
	defer server.Close()
	sent, err := SyncPending(context.Background(), db, server.URL, "4", "2", server.Client())
	if err != nil || sent != 1 || calls.Load() != 2 {
		t.Fatalf("sent=%d calls=%d err=%v", sent, calls.Load(), err)
	}
}

func TestStoreSurvivesReopen(t *testing.T) {
	path := t.TempDir() + "/pos.db"
	db, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Exec(`INSERT INTO pos_operations VALUES ('op-1','1001','150000','PYG','PENDING','2026-09-15T12:00:00Z')`).Error
	if err != nil {
		t.Fatal(err)
	}
	CloseAgentDB(db)
	db, err = OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	defer CloseAgentDB(db)
	var count int
	row, queryErr := agentRow(db, `SELECT COUNT(*) FROM pos_operations WHERE operation_id='op-1'`)
	if err := queryErr; err != nil || row.Scan(&count) != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected one persisted operation, got %d", count)
	}
	err = db.Exec(`INSERT INTO pos_operations VALUES ('op-1','different','1','PYG','PENDING','now')`).Error
	if err == nil {
		t.Fatal("expected duplicate operation_id to fail")
	}
	_ = sql.ErrNoRows
}

func TestStoreRecoversSendingAfterRestart(t *testing.T) {
	path := t.TempDir() + "/recover.db"
	db, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := SaveOperation(db, "op-sending", "1001", "10", "PYG", "SENDING", "2026-09-16"); err != nil {
		t.Fatal(err)
	}
	CloseAgentDB(db)
	db, err = OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	defer CloseAgentDB(db)
	var state string
	row, queryErr := agentRow(db, `SELECT state FROM pos_operations WHERE operation_id='op-sending'`)
	if err := queryErr; err != nil || row.Scan(&state) != nil {
		t.Fatal(err)
	}
	if state != "PENDING" {
		t.Fatalf("expected PENDING after restart, got %s", state)
	}
}

func TestScaleRulesPersistInAgentSQLite(t *testing.T) {
	db, err := OpenStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer CloseAgentDB(db)
	validRules := `[{"prefix":"99","length":10,"product_start":3,"product_length":3,"value_start":6,"value_length":5,"decimals":2,"content":"PRECIO"}]`
	if err := SaveScaleRules(db, "4", validRules, "2026-09-16T12:00:00Z"); err != nil {
		t.Fatal(err)
	}
	rules, err := LoadScaleRules(db, "4")
	if err != nil || rules != validRules {
		t.Fatalf("rules=%s err=%v", rules, err)
	}
	if err := SaveScaleRules(db, "4", "not-json", "now"); err != ErrInvalidOperation {
		t.Fatalf("expected invalid rules rejection, got %v", err)
	}
	if err := SaveScaleRules(db, "4", `{"prefix":"99"}`, "now"); err != ErrInvalidOperation {
		t.Fatalf("expected non-array rules rejection, got %v", err)
	}
	if err := SaveScaleRules(db, "4", `[1]`, "now"); err != ErrInvalidOperation {
		t.Fatalf("expected non-object rule rejection, got %v", err)
	}
	if err := SaveScaleRules(db, "4", `[{"prefix":"99","content":"INVALID"}]`, "now"); err != ErrInvalidOperation {
		t.Fatalf("expected malformed rule rejection, got %v", err)
	}
	if err := db.Exec(`UPDATE pos_scale_rules SET rules_json='not-json' WHERE empresa_id='4'`).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := LoadScaleRules(db, "4"); err != ErrInvalidOperation {
		t.Fatalf("expected invalid cached rules rejection, got %v", err)
	}
}
