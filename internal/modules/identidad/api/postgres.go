package api

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrSessionNotFound = errors.New("session not found")
var ErrInvalidCredentials = errors.New("invalid credentials")

type SessionStore struct{ DB *gorm.DB }

// Authenticate verifica la credencial y el ámbito antes de crear una sesión.
// La comparación se delega a crypt() para no extraer hashes ni contraseñas al proceso.
func (s SessionStore) Authenticate(ctx context.Context, login, password string, companyID, branchID, terminalID int64, now time.Time, ttl time.Duration) (string, error) {
	if s.DB == nil || login == "" || password == "" || companyID <= 0 || branchID <= 0 || terminalID <= 0 {
		return "", ErrInvalidCredentials
	}
	var userID int64
	result := s.DB.WithContext(ctx).Raw(`SELECT u.id FROM erp_v4.gs_usuario u
		JOIN erp_v4.gs_usuario_empresa ue ON ue.usuario_id=u.id AND ue.empresa_id=$3 AND ue.activo
		JOIN erp_v4.gs_usuario_sucursal us ON us.usuario_id=u.id AND us.empresa_id=$3 AND us.sucursal_id=$4 AND us.activo
		JOIN erp_v4.gp_terminal t ON t.empresa_id=$3 AND t.sucursal_id=$4 AND t.id=$5 AND t.activa
		WHERE lower(u.login)=lower($1) AND u.activo AND u.password_hash IS NOT NULL AND u.password_hash=erp_v4.crypt($2,u.password_hash)`, login, password, companyID, branchID, terminalID).Scan(&userID)
	if result.Error != nil {
		return "", result.Error
	}
	if result.RowsAffected == 0 {
		return "", ErrInvalidCredentials
	}
	return s.Create(ctx, userID, companyID, branchID, terminalID, now, ttl)
}

func (s SessionStore) Revoke(ctx context.Context, sessionID string, at time.Time) error {
	if s.DB == nil {
		return ErrInvalidSession
	}
	if _, err := uuid.Parse(sessionID); err != nil {
		return ErrInvalidSession
	}
	result := s.DB.WithContext(ctx).Exec(`UPDATE erp_v4.ypy_usuario_sesion SET revocada_en=$2 WHERE sesion_id=$1 AND revocada_en IS NULL`, sessionID, at)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSessionNotFound
	}
	return nil
}

func (s SessionStore) Create(ctx context.Context, userID, companyID, branchID, terminalID int64, now time.Time, ttl time.Duration) (string, error) {
	if s.DB == nil || userID <= 0 || companyID <= 0 || branchID <= 0 || terminalID <= 0 || ttl <= 0 {
		return "", ErrInvalidSession
	}
	sessionID := uuid.New()
	result := s.DB.WithContext(ctx).Exec(`INSERT INTO erp_v4.ypy_usuario_sesion(sesion_id,usuario_id,empresa_id,sucursal_id,terminal_id,creado_en,expira_en) VALUES($1,$2,$3,$4,$5,$6,$7)`, sessionID, userID, companyID, branchID, terminalID, now, now.Add(ttl))
	if result.Error != nil {
		return "", result.Error
	}
	return sessionID.String(), nil
}

func (s SessionStore) LoadContext(ctx context.Context, sessionID string, now time.Time) (Context, error) {
	if s.DB == nil {
		return Context{}, ErrInvalidSession
	}
	if _, err := uuid.Parse(sessionID); err != nil {
		return Context{}, ErrInvalidSession
	}
	var row struct {
		UserID, CompanyID, BranchID int64
		TerminalID                  *int64
		ExpiresAt                   time.Time
		RevokedAt                   *time.Time
	}
	result := s.DB.WithContext(ctx).Raw(`SELECT usuario_id AS user_id,empresa_id AS company_id,sucursal_id AS branch_id,terminal_id,expira_en AS expires_at,revocada_en AS revoked_at FROM erp_v4.ypy_usuario_sesion WHERE sesion_id=$1`, sessionID).Scan(&row)
	if result.Error != nil {
		return Context{}, result.Error
	}
	if result.RowsAffected == 0 {
		return Context{}, ErrSessionNotFound
	}
	if row.RevokedAt != nil || !now.Before(row.ExpiresAt) || row.TerminalID == nil || row.UserID <= 0 || row.CompanyID <= 0 || row.BranchID <= 0 || *row.TerminalID <= 0 {
		return Context{}, ErrSessionExpired
	}
	var authorized bool
	result = s.DB.WithContext(ctx).Raw(`SELECT EXISTS (
		SELECT 1 FROM erp_v4.gs_usuario u
		JOIN erp_v4.gs_usuario_empresa ue ON ue.usuario_id=u.id AND ue.empresa_id=$2
		JOIN erp_v4.gs_usuario_sucursal us ON us.usuario_id=u.id AND us.empresa_id=$2 AND us.sucursal_id=$3
		JOIN erp_v4.gp_terminal t ON t.empresa_id=$2 AND t.sucursal_id=$3 AND t.id=$4
		WHERE u.id=$1 AND u.activo AND ue.activo AND us.activo AND t.activa
	)`, row.UserID, row.CompanyID, row.BranchID, *row.TerminalID).Scan(&authorized)
	if result.Error != nil {
		return Context{}, result.Error
	}
	if !authorized {
		return Context{}, ErrUnauthorized
	}
	permissions := make(map[string]bool)
	var rows []struct {
		Permission string `gorm:"column:codigo"`
	}
	result = s.DB.WithContext(ctx).Raw(`SELECT DISTINCT p.codigo
		FROM erp_v4.gs_usuario_rol ur
		JOIN erp_v4.gs_rol_permiso rp ON rp.empresa_id=ur.empresa_id AND rp.rol_id=ur.rol_id
		JOIN erp_v4.gs_permiso p ON p.id=rp.permiso_id
		WHERE ur.empresa_id=$1 AND ur.usuario_id=$2 AND (ur.sucursal_id IS NULL OR ur.sucursal_id=$3)`, row.CompanyID, row.UserID, row.BranchID).Scan(&rows)
	if result.Error != nil {
		return Context{}, result.Error
	}
	for _, item := range rows {
		permission := item.Permission
		if permission != "" {
			permissions[permission] = true
		}
	}
	return Context{UserID: strconv.FormatInt(row.UserID, 10), EmpresaID: strconv.FormatInt(row.CompanyID, 10), SucursalID: strconv.FormatInt(row.BranchID, 10), TerminalID: strconv.FormatInt(*row.TerminalID, 10), Active: true, Permissions: permissions}, nil
}
