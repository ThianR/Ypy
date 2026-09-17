BEGIN;
SET LOCAL search_path = erp_v4, public;

CREATE TABLE IF NOT EXISTS gp_sync_cursor (
  empresa_id BIGINT NOT NULL,
  terminal_id BIGINT NOT NULL,
  cursor BIGINT NOT NULL CHECK (cursor >= 0),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (empresa_id, terminal_id),
  FOREIGN KEY (empresa_id, terminal_id) REFERENCES gp_terminal(empresa_id, id)
);

COMMIT;
