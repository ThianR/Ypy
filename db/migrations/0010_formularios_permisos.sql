BEGIN;
SET LOCAL search_path = erp_v4, public;

CREATE TABLE IF NOT EXISTS gs_formulario (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  codigo TEXT NOT NULL UNIQUE,
  nombre TEXT NOT NULL,
  modulo TEXT NOT NULL,
  descripcion TEXT NOT NULL DEFAULT '',
  activo BOOLEAN NOT NULL DEFAULT TRUE,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE gs_permiso ADD COLUMN IF NOT EXISTS activo BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE gs_rol_permiso ALTER COLUMN permiso_id DROP NOT NULL;
ALTER TABLE gs_rol_permiso ADD COLUMN IF NOT EXISTS formulario_id BIGINT REFERENCES gs_formulario(id);
ALTER TABLE gs_rol_permiso ADD COLUMN IF NOT EXISTS usuario_id BIGINT REFERENCES gs_usuario(id);
ALTER TABLE gs_rol_permiso ADD COLUMN IF NOT EXISTS tipo CHAR(1) NOT NULL DEFAULT 'B';
ALTER TABLE gs_rol_permiso ADD COLUMN IF NOT EXISTS visualiza BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE gs_rol_permiso ADD COLUMN IF NOT EXISTS inserta BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE gs_rol_permiso ADD COLUMN IF NOT EXISTS modifica BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE gs_rol_permiso ADD COLUMN IF NOT EXISTS elimina BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE gs_rol_permiso ADD COLUMN IF NOT EXISTS duplica BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE gs_rol_permiso ADD COLUMN IF NOT EXISTS activo BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE gs_rol_permiso ADD CONSTRAINT ck_rol_permiso_formulario_tipo CHECK (
  formulario_id IS NULL OR tipo IN ('B', 'E')
);
ALTER TABLE gs_rol_permiso ADD CONSTRAINT ck_rol_permiso_formulario_destino CHECK (
  formulario_id IS NULL OR ((rol_id IS NOT NULL) <> (usuario_id IS NOT NULL))
);
ALTER TABLE gs_rol_permiso ADD CONSTRAINT ck_rol_permiso_especial CHECK (
  formulario_id IS NULL OR tipo = 'B' OR permiso_id IS NOT NULL
);
ALTER TABLE gs_rol_permiso ADD CONSTRAINT ck_rol_permiso_basico CHECK (
  formulario_id IS NULL OR tipo = 'E' OR permiso_id IS NULL
);

CREATE INDEX IF NOT EXISTS ix_formulario_activo ON gs_formulario (activo, modulo);
CREATE INDEX IF NOT EXISTS ix_rol_permiso_formulario ON gs_rol_permiso (formulario_id, rol_id, usuario_id, activo);

COMMENT ON TABLE gs_formulario IS 'Catálogo estable de formularios visibles y autorizables de Ypy.';
COMMENT ON COLUMN gs_rol_permiso.tipo IS 'B: acciones básicas del formulario; E: permiso especial identificado por permiso_id.';
COMMENT ON COLUMN gs_rol_permiso.permiso_id IS 'Obligatorio solo para asignaciones de tipo E asociadas a un formulario.';

COMMIT;
