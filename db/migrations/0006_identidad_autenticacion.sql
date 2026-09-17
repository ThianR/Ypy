BEGIN;
SET LOCAL search_path = erp_v4, public;

-- La comparación de contraseñas se ejecuta en PostgreSQL mediante crypt().
CREATE EXTENSION IF NOT EXISTS pgcrypto;

COMMIT;
