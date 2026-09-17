BEGIN;
SET LOCAL search_path = erp_v4, public;

CREATE TABLE IF NOT EXISTS ypy_usuario_sesion (
    sesion_id uuid PRIMARY KEY,
    usuario_id bigint NOT NULL,
    empresa_id bigint NOT NULL,
    sucursal_id bigint NOT NULL,
    terminal_id bigint,
    creado_en timestamptz NOT NULL DEFAULT now(),
    expira_en timestamptz NOT NULL,
    revocada_en timestamptz
);
CREATE INDEX IF NOT EXISTS ix_ypy_sesion_usuario ON ypy_usuario_sesion (usuario_id, revocada_en);
COMMENT ON TABLE ypy_usuario_sesion IS 'Sesiones Ypy; validar ámbito contra identidad V4 antes de cada comando';

COMMIT;
