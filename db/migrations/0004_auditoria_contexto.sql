BEGIN;
SET LOCAL search_path = erp_v4, public;

ALTER TABLE gs_auditoria ADD COLUMN IF NOT EXISTS operation_id uuid;
ALTER TABLE gs_auditoria ADD COLUMN IF NOT EXISTS terminal_id bigint;
ALTER TABLE gs_auditoria ADD COLUMN IF NOT EXISTS instalacion_id text;
ALTER TABLE gs_auditoria ADD COLUMN IF NOT EXISTS motivo text;
ALTER TABLE gs_auditoria ADD COLUMN IF NOT EXISTS request_id uuid;
ALTER TABLE gs_auditoria ADD COLUMN IF NOT EXISTS correlation_id uuid;
CREATE INDEX IF NOT EXISTS ix_gs_auditoria_operation ON gs_auditoria (empresa_id, operation_id) WHERE operation_id IS NOT NULL;

COMMIT;

