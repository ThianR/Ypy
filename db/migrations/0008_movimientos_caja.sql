BEGIN;

CREATE TABLE erp_v4.gf_movimiento_caja (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES erp_v4.gs_empresa(id),
  uid UUID NOT NULL UNIQUE,
  apertura_id BIGINT NOT NULL,
  medio_id BIGINT NOT NULL,
  tipo TEXT NOT NULL CHECK (tipo IN ('INGRESO','RETIRO')),
  importe NUMERIC NOT NULL CHECK (importe > 0),
  motivo TEXT NOT NULL CHECK (length(btrim(motivo)) > 0),
  fecha TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (empresa_id, apertura_id) REFERENCES erp_v4.gf_caja_apertura(empresa_id,id),
  FOREIGN KEY (empresa_id, medio_id) REFERENCES erp_v4.gf_medio_pago(empresa_id,id),
  UNIQUE (empresa_id, id)
);

CREATE INDEX gf_movimiento_caja_apertura ON erp_v4.gf_movimiento_caja(empresa_id,apertura_id,medio_id);
COMMIT;
