BEGIN;
SET LOCAL search_path = erp_v4, public;

ALTER TABLE gp_evento_entrada ADD COLUMN IF NOT EXISTS contenido_hash text;
ALTER TABLE gp_evento_entrada ADD COLUMN IF NOT EXISTS resultado jsonb;
ALTER TABLE gp_evento_entrada ADD COLUMN IF NOT EXISTS procesado_en timestamptz;
ALTER TABLE gp_evento_entrada ADD COLUMN IF NOT EXISTS reintentos integer NOT NULL DEFAULT 0 CHECK (reintentos >= 0);
ALTER TABLE gs_evento_salida ADD COLUMN IF NOT EXISTS publicado_error text;
ALTER TABLE gs_evento_salida ADD COLUMN IF NOT EXISTS reintentos integer NOT NULL DEFAULT 0 CHECK (reintentos >= 0);
CREATE INDEX IF NOT EXISTS ix_gp_evento_entrada_pendiente ON gp_evento_entrada (empresa_id, terminal_id, secuencia) WHERE estado = 'PENDIENTE';
CREATE INDEX IF NOT EXISTS ix_gs_evento_salida_pendiente ON gs_evento_salida (empresa_id, creado_en) WHERE publicado_en IS NULL;

COMMIT;
