BEGIN;

CREATE OR REPLACE FUNCTION erp_v4.gf_cerrar_caja(e bigint, apertura bigint, conteos jsonb)
RETURNS void LANGUAGE plpgsql AS $$
DECLARE
  h erp_v4.gf_caja_apertura;
  m record;
  entrada numeric;
  salida numeric;
  contado numeric;
BEGIN
  PERFORM pg_advisory_xact_lock(e);
  SELECT * INTO STRICT h FROM erp_v4.gf_caja_apertura WHERE empresa_id=e AND id=apertura FOR UPDATE;
  IF h.cerrada_en IS NOT NULL THEN RAISE EXCEPTION 'CAJA_YA_CERRADA'; END IF;
  IF EXISTS (SELECT 1 FROM erp_v4.gf_recibo_cabecera WHERE empresa_id=e AND apertura_id=apertura AND estado='BORRADOR') OR
     EXISTS (SELECT 1 FROM erp_v4.gf_pago_cabecera WHERE empresa_id=e AND apertura_id=apertura AND estado='BORRADOR') THEN
    RAISE EXCEPTION 'CAJA_DOCUMENTOS_PENDIENTES';
  END IF;
  FOR m IN SELECT * FROM erp_v4.gf_medio_pago WHERE empresa_id=e LOOP
    SELECT COALESCE(SUM(f.importe),0) INTO entrada
    FROM erp_v4.gf_recibo_forma f
    JOIN erp_v4.gf_recibo_cabecera c ON c.id=f.cabecera_id
    WHERE c.empresa_id=e AND c.apertura_id=apertura AND c.estado='CONFIRMADO' AND f.medio_id=m.id;
    SELECT COALESCE(SUM(f.importe),0) INTO salida
    FROM erp_v4.gf_pago_forma f
    JOIN erp_v4.gf_pago_cabecera c ON c.id=f.cabecera_id
    WHERE c.empresa_id=e AND c.apertura_id=apertura AND c.estado='CONFIRMADO' AND f.medio_id=m.id;
    SELECT entrada + COALESCE(SUM(CASE WHEN tipo='INGRESO' THEN importe ELSE 0 END),0),
           salida + COALESCE(SUM(CASE WHEN tipo='RETIRO' THEN importe ELSE 0 END),0)
    INTO entrada, salida
    FROM erp_v4.gf_movimiento_caja
    WHERE empresa_id=e AND apertura_id=apertura AND medio_id=m.id;
    IF m.efectivo THEN entrada:=entrada+h.fondo; END IF;
    contado:=(conteos->>m.id::text)::numeric;
    IF contado IS NULL OR contado<0 THEN RAISE EXCEPTION 'FALTA_CONTEO_MEDIO'; END IF;
    INSERT INTO erp_v4.gf_caja_arqueo(empresa_id,apertura_id,medio_id,contado,esperado)
    VALUES(e,apertura,m.id,contado,entrada-salida);
  END LOOP;
  UPDATE erp_v4.gf_caja_apertura SET cerrada_en=clock_timestamp() WHERE id=apertura;
END;
$$;

ALTER FUNCTION erp_v4.gf_cerrar_caja(bigint,bigint,jsonb) SET search_path=erp_v4,pg_catalog;
COMMIT;
