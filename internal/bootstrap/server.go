package bootstrap

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	catalogapi "github.com/ypy-erp/ypy/internal/modules/catalogo/api"
	identity "github.com/ypy-erp/ypy/internal/modules/identidad/api"
	maestrosapi "github.com/ypy-erp/ypy/internal/modules/maestros/api"
	treasuryapi "github.com/ypy-erp/ypy/internal/modules/tesoreria/api"
	ventasapi "github.com/ypy-erp/ypy/internal/modules/ventas/api"
	"github.com/ypy-erp/ypy/internal/platform/httpx"
	"github.com/ypy-erp/ypy/internal/poscentral"
	"gorm.io/gorm"
	"io"
	"mime"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type operationStore interface {
	Apply(poscentral.Operation) (poscentral.Operation, error)
}
type contextOperationStore interface {
	ApplyContext(context.Context, poscentral.Operation) (poscentral.Operation, error)
}

func isJSONContentType(value string) bool {
	mediaType, _, err := mime.ParseMediaType(value)
	return err == nil && mediaType == "application/json"
}

func sessionAllowsScope(r *http.Request, empresaID, terminalID string) bool {
	if os.Getenv("YPY_REQUIRE_SESSION") != "1" {
		return true
	}
	ctx, ok := identity.RequestContext(r.Context())
	return ok && ctx.EmpresaID == empresaID && ctx.TerminalID == terminalID
}

func sessionAllowsCompany(r *http.Request, empresaID string) bool {
	if os.Getenv("YPY_REQUIRE_SESSION") != "1" {
		return true
	}
	ctx, ok := identity.RequestContext(r.Context())
	return ok && ctx.EmpresaID == empresaID
}

func BuildRouter(db *sql.DB, gormDB *gorm.DB) http.Handler {
	mux := http.NewServeMux()
	var store operationStore = poscentral.NewStore()
	if gormDB != nil {
		store = poscentral.PostgresStore{DB: gormDB}
	}
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok","service":"ypy-erp"}`)
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if db != nil {
			if err := db.PingContext(r.Context()); err != nil {
				httpx.WriteError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "database unavailable", httpx.RequestID(r.Context()), true)
				return
			}
		}
		fmt.Fprint(w, `{"status":"ready","service":"ypy-erp"}`)
	})
	var sessionManager identity.SessionManager
	if gormDB != nil {
		sessionManager = identity.SessionStore{DB: gormDB}
	}
	mux.Handle("/api/v1/auth/session", identity.SessionHandler(sessionManager))
	mux.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || db == nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, "AUTH_UNAVAILABLE", "authentication requires PostgreSQL", httpx.RequestID(r.Context()), true)
			return
		}
		var request struct {
			Login      string `json:"login"`
			Password   string `json:"password"`
			EmpresaID  int64  `json:"empresa_id"`
			SucursalID int64  `json:"sucursal_id"`
			TerminalID int64  `json:"terminal_id"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
		if !isJSONContentType(r.Header.Get("Content-Type")) || json.NewDecoder(r.Body).Decode(&request) != nil {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_LOGIN", "invalid login request", httpx.RequestID(r.Context()), false)
			return
		}
		sessionID, err := (identity.SessionStore{DB: gormDB}).Authenticate(r.Context(), request.Login, request.Password, request.EmpresaID, request.SucursalID, request.TerminalID, time.Now().UTC(), 12*time.Hour)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid credentials or scope", httpx.RequestID(r.Context()), false)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"session_id": sessionID, "status": "ACTIVE"})
	})
	mux.HandleFunc("/api/v1/maestros/unidades-medida", func(w http.ResponseWriter, r *http.Request) {
		if gormDB == nil { httpx.WriteError(w, 503, "DATABASE_UNAVAILABLE", "base de datos no disponible", httpx.RequestID(r.Context()), true); return }
		if r.Method == http.MethodGet { rows := make([]struct { ID int64 `json:"id"`; Codigo string `json:"codigo"`; Nombre string `json:"nombre"`; Dimension string `json:"dimension"`; AdmiteFraccion bool `json:"admiteFraccion"` }, 0); if err := gormDB.WithContext(r.Context()).Raw(`SELECT id,codigo,nombre,dimension,admite_fraccion FROM erp_v4.gi_unidad_medida ORDER BY codigo`).Scan(&rows).Error; err != nil { httpx.WriteError(w, 500, "DATABASE_ERROR", "no se pudieron consultar las unidades", httpx.RequestID(r.Context()), true); return }; w.Header().Set("Content-Type", "application/json"); _ = json.NewEncoder(w).Encode(rows); return }
		if r.Method == http.MethodPost || r.Method == http.MethodPut { var input struct { ID int64 `json:"id"`; Codigo string `json:"codigo"`; Nombre string `json:"nombre"`; Dimension string `json:"dimension"`; AdmiteFraccion bool `json:"admiteFraccion"` }; r.Body = http.MaxBytesReader(w, r.Body, 32<<10); if json.NewDecoder(r.Body).Decode(&input) != nil || input.Codigo == "" || input.Nombre == "" || input.Dimension == "" { httpx.WriteError(w, 400, "INVALID_UNIT", "código, nombre y dimensión son obligatorios", httpx.RequestID(r.Context()), false); return }; input.Codigo = strings.ToUpper(strings.TrimSpace(input.Codigo)); input.Nombre = strings.TrimSpace(input.Nombre); input.Dimension = strings.ToUpper(strings.TrimSpace(input.Dimension)); var err error; if r.Method == http.MethodPost { err = gormDB.WithContext(r.Context()).Exec(`INSERT INTO erp_v4.gi_unidad_medida(codigo,nombre,dimension,admite_fraccion) VALUES(?,?,?,?)`, input.Codigo,input.Nombre,input.Dimension,input.AdmiteFraccion).Error } else { if input.ID <= 0 { httpx.WriteError(w,400,"INVALID_ID","identificador inválido",httpx.RequestID(r.Context()),false); return }; err = gormDB.WithContext(r.Context()).Exec(`UPDATE erp_v4.gi_unidad_medida SET codigo=?,nombre=?,dimension=?,admite_fraccion=? WHERE id=?`, input.Codigo,input.Nombre,input.Dimension,input.AdmiteFraccion,input.ID).Error }; if err != nil { httpx.WriteError(w,409,"UNIT_SAVE_FAILED","no se pudo guardar la unidad",httpx.RequestID(r.Context()),false); return }; w.Header().Set("Content-Type","application/json"); _ = json.NewEncoder(w).Encode(map[string]bool{"ok":true}); return }
		httpx.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", httpx.RequestID(r.Context()), false)
	})
	mux.HandleFunc("/api/v1/maestros/unidades-medida/", func(w http.ResponseWriter, r *http.Request) { if gormDB == nil || r.Method != http.MethodDelete { httpx.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", httpx.RequestID(r.Context()), false); return }; id, err := strconv.ParseInt(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/maestros/unidades-medida/"), "/"), 10, 64); if err != nil || id <= 0 { httpx.WriteError(w,400,"INVALID_ID","identificador inválido",httpx.RequestID(r.Context()),false); return }; result := gormDB.WithContext(r.Context()).Exec(`DELETE FROM erp_v4.gi_unidad_medida WHERE id=?`, id); if result.Error != nil || result.RowsAffected == 0 { httpx.WriteError(w,404,"UNIT_NOT_FOUND","la unidad no existe",httpx.RequestID(r.Context()),false); return }; w.Header().Set("Content-Type","application/json"); _ = json.NewEncoder(w).Encode(map[string]bool{"ok":true}) })
	mux.HandleFunc("/api/v1/maestros/categorias", func(w http.ResponseWriter, r *http.Request) {
		if gormDB == nil || r.Method != http.MethodGet { httpx.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", httpx.RequestID(r.Context()), false); return }
		companyID := "1"; if ctx, ok := identity.RequestContext(r.Context()); ok && ctx.EmpresaID != "" { companyID = ctx.EmpresaID }
		rows := make([]struct { ID int64 `json:"id"`; Codigo string `json:"codigo"`; Nombre string `json:"nombre"`; CreatedAt time.Time `json:"createdAt"`; UpdatedAt time.Time `json:"updatedAt"` }, 0)
		if err := gormDB.WithContext(r.Context()).Raw(`SELECT id,codigo,nombre,now() AS created_at,now() AS updated_at FROM erp_v4.gi_categoria WHERE empresa_id=$1 ORDER BY codigo`, companyID).Scan(&rows).Error; err != nil { httpx.WriteError(w, 500, "DATABASE_ERROR", "no se pudieron consultar las categorías", httpx.RequestID(r.Context()), true); return }
		w.Header().Set("Content-Type", "application/json"); _ = json.NewEncoder(w).Encode(rows)
	})
	mux.HandleFunc("/api/v1/pos/operations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			httpx.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", httpx.RequestID(r.Context()), false)
			return
		}
		if !isJSONContentType(r.Header.Get("Content-Type")) {
			httpx.WriteError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "content type must be application/json", httpx.RequestID(r.Context()), false)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			httpx.WriteError(w, 400, "INVALID_BODY", "invalid body", httpx.RequestID(r.Context()), false)
			return
		}
		op, err := poscentral.Decode(body)
		if err != nil {
			httpx.WriteError(w, 400, "INVALID_OPERATION", "invalid operation", httpx.RequestID(r.Context()), false)
			return
		}
		if !sessionAllowsScope(r, op.EmpresaID, op.TerminalID) {
			httpx.WriteError(w, http.StatusForbidden, "SCOPE_FORBIDDEN", "operation scope is not authorized", httpx.RequestID(r.Context()), false)
			return
		}
		var result poscentral.Operation
		if contextual, ok := store.(contextOperationStore); ok {
			result, err = contextual.ApplyContext(r.Context(), op)
		} else {
			result, err = store.Apply(op)
		}
		if err == poscentral.ErrConflict {
			httpx.WriteError(w, http.StatusConflict, "OPERATION_CONTENT_CONFLICT", "operation content conflict", httpx.RequestID(r.Context()), false)
			return
		}
		if err != nil {
			httpx.WriteError(w, 500, "APPLY_FAILED", "apply failed", httpx.RequestID(r.Context()), true)
			return
		}
		if pg, ok := store.(poscentral.PostgresStore); ok {
			if err := pg.EnqueueAccepted(r.Context(), result, time.Now().UTC()); err != nil {
				httpx.WriteError(w, http.StatusInternalServerError, "EVENT_ENQUEUE_FAILED", "accepted event could not be queued", httpx.RequestID(r.Context()), true)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"operation_id": result.ID, "content_hash": poscentral.ContentHash(result.Payload), "status": "ACCEPTED"})
	})
	mux.HandleFunc("/api/v1/pos/sales", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			httpx.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", httpx.RequestID(r.Context()), false)
			return
		}
		if !isJSONContentType(r.Header.Get("Content-Type")) {
			httpx.WriteError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "content type must be application/json", httpx.RequestID(r.Context()), false)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		pg, ok := store.(poscentral.PostgresStore)
		if !ok {
			httpx.WriteError(w, http.StatusServiceUnavailable, "DATABASE_REQUIRED", "commercial sale integration requires PostgreSQL", httpx.RequestID(r.Context()), true)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid body", httpx.RequestID(r.Context()), false)
			return
		}
		op, err := poscentral.Decode(body)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_OPERATION", "invalid operation", httpx.RequestID(r.Context()), false)
			return
		}
		if !sessionAllowsScope(r, op.EmpresaID, op.TerminalID) {
			httpx.WriteError(w, http.StatusForbidden, "SCOPE_FORBIDDEN", "operation scope is not authorized", httpx.RequestID(r.Context()), false)
			return
		}
		var salePayload struct {
			Forms []json.RawMessage `json:"formas"`
		}
		if err := json.Unmarshal([]byte(op.Payload), &salePayload); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_OPERATION", "invalid sale payload", httpx.RequestID(r.Context()), false)
			return
		}
		var saleID int64
		if len(salePayload.Forms) > 0 {
			saleID, err = pg.IntegrateSaleWithReceipt(r.Context(), op)
		} else {
			saleID, err = pg.IntegrateSale(r.Context(), op)
		}
		if err != nil {
			httpx.WriteError(w, http.StatusConflict, "SALE_INTEGRATION_FAILED", err.Error(), httpx.RequestID(r.Context()), false)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"operation_id": op.ID, "sale_id": saleID, "status": "APPLIED"})
	})
	mux.HandleFunc("/api/v1/sales/returns", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || db == nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, "SALES_UNAVAILABLE", "sales require PostgreSQL", httpx.RequestID(r.Context()), true)
			return
		}
		if !isJSONContentType(r.Header.Get("Content-Type")) {
			httpx.WriteError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "content type must be application/json", httpx.RequestID(r.Context()), false)
			return
		}
		var request struct {
			ID        string `json:"id"`
			CompanyID int64  `json:"empresa_id"`
			LineID    int64  `json:"linea_id"`
			Quantity  string `json:"cantidad"`
			Reason    string `json:"motivo"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_RETURN", "invalid return request", httpx.RequestID(r.Context()), false)
			return
		}
		id, err := uuid.Parse(request.ID)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_RETURN", "invalid return identifier", httpx.RequestID(r.Context()), false)
			return
		}
		if !sessionAllowsCompany(r, strconv.FormatInt(request.CompanyID, 10)) {
			httpx.WriteError(w, http.StatusForbidden, "SCOPE_FORBIDDEN", "return scope is not authorized", httpx.RequestID(r.Context()), false)
			return
		}
		returnID, err := (&ventasapi.ReturnStore{DB: gormDB}).Create(r.Context(), request.CompanyID, request.LineID, request.Quantity, request.Reason, id)
		if err != nil {
			httpx.WriteError(w, http.StatusConflict, "RETURN_FAILED", "return could not be registered", httpx.RequestID(r.Context()), false)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"return_id": returnID, "status": "REGISTERED"})
	})
	mux.HandleFunc("/api/v1/sales/credit-notes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || db == nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, "SALES_UNAVAILABLE", "sales require PostgreSQL", httpx.RequestID(r.Context()), true)
			return
		}
		if !isJSONContentType(r.Header.Get("Content-Type")) {
			httpx.WriteError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "content type must be application/json", httpx.RequestID(r.Context()), false)
			return
		}
		var n ventasapi.CreditNote
		r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
		if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_CREDIT_NOTE", "invalid credit note", httpx.RequestID(r.Context()), false)
			return
		}
		if !sessionAllowsCompany(r, strconv.FormatInt(n.CompanyID, 10)) {
			httpx.WriteError(w, http.StatusForbidden, "SCOPE_FORBIDDEN", "credit note scope is not authorized", httpx.RequestID(r.Context()), false)
			return
		}
		id, err := (&ventasapi.CreditNoteStore{DB: gormDB}).Create(r.Context(), n)
		if err == nil {
			err = (&ventasapi.CreditNoteStore{DB: gormDB}).Confirm(r.Context(), n.CompanyID, id)
		}
		if err != nil {
			httpx.WriteError(w, http.StatusConflict, "CREDIT_NOTE_FAILED", "credit note could not be confirmed", httpx.RequestID(r.Context()), false)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"credit_note_id": id, "status": "CONFIRMED"})
	})
	mux.HandleFunc("/api/v1/sales/credits/apply", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || db == nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, "TREASURY_UNAVAILABLE", "treasury requires PostgreSQL", httpx.RequestID(r.Context()), true)
			return
		}
		if !isJSONContentType(r.Header.Get("Content-Type")) {
			httpx.WriteError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "content type must be application/json", httpx.RequestID(r.Context()), false)
			return
		}
		var request struct {
			ID            string `json:"id"`
			CompanyID     int64  `json:"empresa_id"`
			CreditID      int64  `json:"credito_id"`
			InstallmentID int64  `json:"cuota_id"`
			Amount        string `json:"importe"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_CREDIT_APPLICATION", "invalid credit application", httpx.RequestID(r.Context()), false)
			return
		}
		id, err := uuid.Parse(request.ID)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_CREDIT_APPLICATION", "invalid application identifier", httpx.RequestID(r.Context()), false)
			return
		}
		if !sessionAllowsCompany(r, strconv.FormatInt(request.CompanyID, 10)) {
			httpx.WriteError(w, http.StatusForbidden, "SCOPE_FORBIDDEN", "credit scope is not authorized", httpx.RequestID(r.Context()), false)
			return
		}
		applicationID, err := (&treasuryapi.CreditStore{DB: gormDB}).Apply(r.Context(), request.CompanyID, request.CreditID, request.InstallmentID, request.Amount, id)
		if err != nil {
			httpx.WriteError(w, http.StatusConflict, "CREDIT_APPLICATION_FAILED", "credit could not be applied", httpx.RequestID(r.Context()), false)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"application_id": applicationID, "status": "APPLIED"})
	})
	mux.HandleFunc("/api/v1/pos/sync/cursor", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			httpx.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", httpx.RequestID(r.Context()), false)
			return
		}
		pg, ok := store.(poscentral.PostgresStore)
		if !ok {
			httpx.WriteError(w, http.StatusServiceUnavailable, "DATABASE_REQUIRED", "commercial synchronization requires PostgreSQL", httpx.RequestID(r.Context()), true)
			return
		}
		company, err1 := strconv.ParseInt(r.URL.Query().Get("empresa_id"), 10, 64)
		terminal, err2 := strconv.ParseInt(r.URL.Query().Get("terminal_id"), 10, 64)
		if err1 != nil || err2 != nil || company <= 0 || terminal <= 0 {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_SCOPE", "empresa_id and terminal_id are required", httpx.RequestID(r.Context()), false)
			return
		}
		if !sessionAllowsScope(r, strconv.FormatInt(company, 10), strconv.FormatInt(terminal, 10)) {
			httpx.WriteError(w, http.StatusForbidden, "SCOPE_FORBIDDEN", "cursor scope is not authorized", httpx.RequestID(r.Context()), false)
			return
		}
		cursor, err := pg.Cursor(r.Context(), company, terminal)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "CURSOR_READ_FAILED", "cursor read failed", httpx.RequestID(r.Context()), true)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"empresa_id": company, "terminal_id": terminal, "cursor": strconv.FormatInt(cursor, 10)})
	})
	mux.HandleFunc("/api/v1/catalog/scale-rules", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			httpx.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", httpx.RequestID(r.Context()), false)
			return
		}
		if db == nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, "DATABASE_REQUIRED", "scale rules require PostgreSQL", httpx.RequestID(r.Context()), true)
			return
		}
		company, err := strconv.ParseInt(r.URL.Query().Get("empresa_id"), 10, 64)
		if err != nil || company <= 0 {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_SCOPE", "empresa_id is required", httpx.RequestID(r.Context()), false)
			return
		}
		rules, err := catalogapi.LoadV4ScaleRules(r.Context(), gormDB, company)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "SCALE_RULES_READ_FAILED", "scale rules read failed", httpx.RequestID(r.Context()), true)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"empresa_id": company, "rules": rules})
	})
	mux.HandleFunc("/api/v1/catalog/barcodes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || db == nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, "CATALOG_UNAVAILABLE", "catalog requires PostgreSQL", httpx.RequestID(r.Context()), true)
			return
		}
		company, err := strconv.ParseInt(r.URL.Query().Get("empresa_id"), 10, 64)
		code := strings.TrimSpace(r.URL.Query().Get("codigo"))
		if err != nil || company <= 0 || code == "" {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_BARCODE_QUERY", "invalid barcode query", httpx.RequestID(r.Context()), false)
			return
		}
		if os.Getenv("YPY_REQUIRE_SESSION") == "1" {
			ctx, ok := identity.RequestContext(r.Context())
			if !ok || ctx.EmpresaID != strconv.FormatInt(company, 10) {
				httpx.WriteError(w, http.StatusForbidden, "SCOPE_FORBIDDEN", "company scope forbidden", httpx.RequestID(r.Context()), false)
				return
			}
		}
		match, err := (&catalogapi.BarcodeStore{DB: gormDB}).Resolve(r.Context(), company, code)
		if err != nil {
			httpx.WriteError(w, http.StatusNotFound, "BARCODE_NOT_FOUND", "barcode not found", httpx.RequestID(r.Context()), false)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(match)
	})
	mux.HandleFunc("/api/v1/catalog/items", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || db == nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, "CATALOG_UNAVAILABLE", "catalog requires PostgreSQL", httpx.RequestID(r.Context()), true)
			return
		}
		q := r.URL.Query()
		company, e1 := strconv.ParseInt(q.Get("empresa_id"), 10, 64)
		list, e2 := strconv.ParseInt(q.Get("lista_id"), 10, 64)
		at, e3 := time.Parse(time.RFC3339, q.Get("as_of"))
		query := strings.TrimSpace(q.Get("q"))
		limit, e4 := strconv.Atoi(q.Get("limite"))
		if limit == 0 {
			limit = 25
		}
		if e1 != nil || e2 != nil || e3 != nil || e4 != nil || company <= 0 || list <= 0 || query == "" {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_CATALOG_QUERY", "invalid catalog query", httpx.RequestID(r.Context()), false)
			return
		}
		if !sessionAllowsCompany(r, strconv.FormatInt(company, 10)) {
			httpx.WriteError(w, http.StatusForbidden, "SCOPE_FORBIDDEN", "company scope forbidden", httpx.RequestID(r.Context()), false)
			return
		}
		items, err := (&catalogapi.BarcodeStore{DB: gormDB}).Search(r.Context(), company, list, query, at, limit)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "CATALOG_READ_FAILED", "catalog search failed", httpx.RequestID(r.Context()), true)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"empresa_id": company, "items": items})
	})
	mux.HandleFunc("/api/v1/catalog/prices", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || db == nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, "CATALOG_UNAVAILABLE", "catalog requires PostgreSQL", httpx.RequestID(r.Context()), true)
			return
		}
		q := r.URL.Query()
		company, e1 := strconv.ParseInt(q.Get("empresa_id"), 10, 64)
		list, e2 := strconv.ParseInt(q.Get("lista_id"), 10, 64)
		presentation, e3 := strconv.ParseInt(q.Get("presentacion_id"), 10, 64)
		at, e4 := time.Parse(time.RFC3339, q.Get("as_of"))
		quantity := q.Get("cantidad")
		if e1 != nil || e2 != nil || e3 != nil || e4 != nil || company <= 0 || list <= 0 || presentation <= 0 || quantity == "" {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_PRICE_QUERY", "invalid price query", httpx.RequestID(r.Context()), false)
			return
		}
		if os.Getenv("YPY_REQUIRE_SESSION") == "1" {
			ctx, ok := identity.RequestContext(r.Context())
			if !ok || ctx.EmpresaID != strconv.FormatInt(company, 10) {
				httpx.WriteError(w, http.StatusForbidden, "SCOPE_FORBIDDEN", "company scope forbidden", httpx.RequestID(r.Context()), false)
				return
			}
		}
		price, err := (&catalogapi.BarcodeStore{DB: gormDB}).Price(r.Context(), company, list, presentation, quantity, at)
		if err != nil {
			httpx.WriteError(w, http.StatusNotFound, "PRICE_NOT_FOUND", "effective price not found", httpx.RequestID(r.Context()), false)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"empresa_id": strconv.FormatInt(company, 10), "precio": price, "as_of": at.Format(time.RFC3339)})
	})
	mux.HandleFunc("/api/v1/cash/open", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || db == nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, "CASH_UNAVAILABLE", "cash requires PostgreSQL", httpx.RequestID(r.Context()), true)
			return
		}
		var req struct {
			EmpresaID  int64  `json:"empresa_id"`
			CajaID     int64  `json:"caja_id"`
			UsuarioID  int64  `json:"usuario_id"`
			TerminalID int64  `json:"terminal_id"`
			Fondo      string `json:"fondo"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil || req.EmpresaID <= 0 || req.CajaID <= 0 || req.UsuarioID <= 0 || req.TerminalID <= 0 || req.Fondo == "" {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_CASH_OPEN", "invalid cash opening", httpx.RequestID(r.Context()), false)
			return
		}
		if os.Getenv("YPY_REQUIRE_SESSION") == "1" {
			ctx, ok := identity.RequestContext(r.Context())
			if !ok || ctx.EmpresaID != strconv.FormatInt(req.EmpresaID, 10) || ctx.UserID != strconv.FormatInt(req.UsuarioID, 10) || ctx.TerminalID != strconv.FormatInt(req.TerminalID, 10) {
				httpx.WriteError(w, http.StatusForbidden, "SCOPE_FORBIDDEN", "company scope forbidden", httpx.RequestID(r.Context()), false)
				return
			}
		}
		id, err := (&treasuryapi.PostgresStore{DB: gormDB}).Open(r.Context(), req.EmpresaID, req.CajaID, req.UsuarioID, req.TerminalID, req.Fondo)
		if err != nil {
			httpx.WriteError(w, http.StatusConflict, "CASH_OPEN_FAILED", "cash opening failed", httpx.RequestID(r.Context()), true)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int64{"apertura_id": id})
	})
	mux.HandleFunc("/api/v1/cash/receipts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			httpx.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", httpx.RequestID(r.Context()), false)
			return
		}
		if db == nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, "CASH_UNAVAILABLE", "cash requires PostgreSQL", httpx.RequestID(r.Context()), true)
			return
		}
		if !isJSONContentType(r.Header.Get("Content-Type")) {
			httpx.WriteError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "content type must be application/json", httpx.RequestID(r.Context()), false)
			return
		}
		var request struct {
			ID         string                    `json:"id"`
			CompanyID  int64                     `json:"empresa_id"`
			CustomerID int64                     `json:"cliente_id"`
			OpeningID  int64                     `json:"apertura_id"`
			CurrencyID int64                     `json:"moneda_id"`
			Amount     string                    `json:"importe"`
			Date       string                    `json:"fecha"`
			Forms      []treasuryapi.ReceiptForm `json:"formas"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 256<<10)
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_RECEIPT", "invalid receipt request", httpx.RequestID(r.Context()), false)
			return
		}
		id, err := uuid.Parse(request.ID)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_RECEIPT", "invalid receipt identifier", httpx.RequestID(r.Context()), false)
			return
		}
		if !sessionAllowsCompany(r, strconv.FormatInt(request.CompanyID, 10)) {
			httpx.WriteError(w, http.StatusForbidden, "SCOPE_FORBIDDEN", "receipt scope is not authorized", httpx.RequestID(r.Context()), false)
			return
		}
		for i := range request.Forms {
			request.Forms[i].CompanyID = request.CompanyID
			request.Forms[i].HeaderID = 0
		}
		headerID, err := (&treasuryapi.ReceiptStore{DB: gormDB}).CreateAndConfirm(r.Context(), treasuryapi.Receipt{ID: id, CompanyID: request.CompanyID, CustomerID: request.CustomerID, OpeningID: request.OpeningID, CurrencyID: request.CurrencyID, Amount: request.Amount, Date: request.Date}, request.Forms)
		if err != nil {
			httpx.WriteError(w, http.StatusConflict, "RECEIPT_FAILED", "receipt could not be created or confirmed", httpx.RequestID(r.Context()), false)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"receipt_id": headerID, "status": "CONFIRMED"})
	})
	mux.HandleFunc("/api/v1/cash/payments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || db == nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, "CASH_UNAVAILABLE", "cash requires PostgreSQL", httpx.RequestID(r.Context()), true)
			return
		}
		if !isJSONContentType(r.Header.Get("Content-Type")) {
			httpx.WriteError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "content type must be application/json", httpx.RequestID(r.Context()), false)
			return
		}
		var request struct {
			Payment treasuryapi.PaymentDocument `json:"pago"`
			Forms   []treasuryapi.PaymentForm   `json:"formas"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 256<<10)
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_PAYMENT", "invalid payment request", httpx.RequestID(r.Context()), false)
			return
		}
		if !sessionAllowsCompany(r, strconv.FormatInt(request.Payment.CompanyID, 10)) {
			httpx.WriteError(w, http.StatusForbidden, "SCOPE_FORBIDDEN", "payment scope is not authorized", httpx.RequestID(r.Context()), false)
			return
		}
		id, err := (&treasuryapi.PaymentStore{DB: gormDB}).CreateAndConfirm(r.Context(), request.Payment, request.Forms)
		if err != nil {
			httpx.WriteError(w, http.StatusConflict, "PAYMENT_FAILED", "payment could not be confirmed", httpx.RequestID(r.Context()), false)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"payment_id": id, "status": "CONFIRMED"})
	})
	mux.HandleFunc("/api/v1/cash/movements", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			httpx.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", httpx.RequestID(r.Context()), false)
			return
		}
		if db == nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, "CASH_UNAVAILABLE", "cash requires PostgreSQL", httpx.RequestID(r.Context()), true)
			return
		}
		if !isJSONContentType(r.Header.Get("Content-Type")) {
			httpx.WriteError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "content type must be application/json", httpx.RequestID(r.Context()), false)
			return
		}
		var request struct {
			ID        string `json:"id"`
			CompanyID int64  `json:"empresa_id"`
			OpeningID int64  `json:"apertura_id"`
			MethodID  int64  `json:"medio_id"`
			Kind      string `json:"tipo"`
			Amount    string `json:"importe"`
			Reason    string `json:"motivo"`
			Date      string `json:"fecha"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_CASH_MOVEMENT", "invalid cash movement", httpx.RequestID(r.Context()), false)
			return
		}
		id, err := uuid.Parse(request.ID)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_CASH_MOVEMENT", "invalid movement identifier", httpx.RequestID(r.Context()), false)
			return
		}
		if !sessionAllowsCompany(r, strconv.FormatInt(request.CompanyID, 10)) {
			httpx.WriteError(w, http.StatusForbidden, "SCOPE_FORBIDDEN", "movement scope is not authorized", httpx.RequestID(r.Context()), false)
			return
		}
		movementID, err := (&treasuryapi.PostgresStore{DB: gormDB}).AddMovement(r.Context(), treasuryapi.CashMovement{ID: id, CompanyID: request.CompanyID, OpeningID: request.OpeningID, MethodID: request.MethodID, Kind: request.Kind, Amount: request.Amount, Reason: request.Reason, Date: request.Date})
		if err != nil {
			httpx.WriteError(w, http.StatusConflict, "CASH_MOVEMENT_FAILED", "cash movement could not be recorded", httpx.RequestID(r.Context()), false)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"movement_id": movementID, "status": "RECORDED"})
	})
	mux.HandleFunc("/api/v1/cash/close", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || db == nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, "CASH_UNAVAILABLE", "cash requires PostgreSQL", httpx.RequestID(r.Context()), true)
			return
		}
		var req struct {
			EmpresaID  int64           `json:"empresa_id"`
			AperturaID int64           `json:"apertura_id"`
			Conteos    json.RawMessage `json:"conteos"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil || req.EmpresaID <= 0 || req.AperturaID <= 0 || !json.Valid(req.Conteos) {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_CASH_CLOSE", "invalid cash closing", httpx.RequestID(r.Context()), false)
			return
		}
		if os.Getenv("YPY_REQUIRE_SESSION") == "1" {
			ctx, ok := identity.RequestContext(r.Context())
			if !ok || ctx.EmpresaID != strconv.FormatInt(req.EmpresaID, 10) {
				httpx.WriteError(w, http.StatusForbidden, "SCOPE_FORBIDDEN", "company scope forbidden", httpx.RequestID(r.Context()), false)
				return
			}
		}
		if err := (&treasuryapi.PostgresStore{DB: gormDB}).Close(r.Context(), req.EmpresaID, req.AperturaID, req.Conteos); err != nil {
			httpx.WriteError(w, http.StatusConflict, "CASH_CLOSE_FAILED", "cash closing failed", httpx.RequestID(r.Context()), true)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"CLOSED"}`))
	})
	mux.HandleFunc("/api/v1/maestros/quotes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || db == nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, "QUOTES_UNAVAILABLE", "quotes require PostgreSQL", httpx.RequestID(r.Context()), true)
			return
		}
		q := r.URL.Query()
		company, e1 := strconv.ParseInt(q.Get("empresa_id"), 10, 64)
		from, e2 := strconv.ParseInt(q.Get("moneda_origen_id"), 10, 64)
		to, e3 := strconv.ParseInt(q.Get("moneda_destino_id"), 10, 64)
		at, e4 := time.Parse(time.RFC3339, q.Get("as_of"))
		if e1 != nil || e2 != nil || e3 != nil || e4 != nil || company <= 0 || from <= 0 || to <= 0 {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_QUOTE_QUERY", "invalid quote query", httpx.RequestID(r.Context()), false)
			return
		}
		if os.Getenv("YPY_REQUIRE_SESSION") == "1" {
			ctx, ok := identity.RequestContext(r.Context())
			if !ok || ctx.EmpresaID != strconv.FormatInt(company, 10) {
				httpx.WriteError(w, http.StatusForbidden, "SCOPE_FORBIDDEN", "company scope forbidden", httpx.RequestID(r.Context()), false)
				return
			}
		}
		quote, err := (&maestrosapi.QuoteStore{DB: gormDB}).LoadQuote(r.Context(), company, from, to, q.Get("tipo"), at)
		if err != nil {
			httpx.WriteError(w, http.StatusNotFound, "QUOTE_NOT_FOUND", "effective quote not found", httpx.RequestID(r.Context()), false)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(quote)
	})

	var handler http.Handler = httpx.WithRequestID(mux)
	if os.Getenv("YPY_REQUIRE_SESSION") == "1" && db != nil {
		protected := identity.WithSession(mux, identity.SessionStore{DB: gormDB})
		handler = httpx.WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") && r.URL.Path != "/api/v1/auth/login" {
				protected.ServeHTTP(w, r)
				return
			}
			mux.ServeHTTP(w, r)
		}))
	}
	return handler
}
