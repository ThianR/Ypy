BEGIN;

CREATE OR REPLACE FUNCTION erp_v4.gp_integrar_venta_cobro(e bigint, terminal bigint, secuencia bigint, clave uuid, ocurrido timestamptz, datos jsonb)
RETURNS bigint LANGUAGE plpgsql AS $$
DECLARE
  venta bigint;
  recibo bigint;
  forma jsonb;
  total numeric;
  recibo_numero bigint;
  recibo_clave uuid;
BEGIN
  venta := erp_v4.gp_integrar_venta(e, terminal, secuencia, clave, ocurrido, datos);
  IF venta IS NULL THEN
    RAISE EXCEPTION 'VENTA_NO_APLICADA';
  END IF;

  SELECT id INTO recibo FROM erp_v4.gf_recibo_cabecera WHERE empresa_id=e AND uid=clave FOR UPDATE;
  IF recibo IS NOT NULL THEN
    IF NOT EXISTS (SELECT 1 FROM erp_v4.gf_recibo_cabecera WHERE empresa_id=e AND id=recibo AND estado='CONFIRMADO') THEN
      RAISE EXCEPTION 'RECIBO_NO_CONFIRMADO';
    END IF;
    RETURN venta;
  END IF;

  IF (datos->>'apertura_id') IS NULL OR (datos->>'cliente_id') IS NULL OR (datos->>'moneda_id') IS NULL OR (datos->>'total') IS NULL OR (datos->>'usuario_id') IS NULL OR (datos->>'recibo_uid') IS NULL OR (datos->>'recibo_talonario_id') IS NULL OR (datos->>'recibo_rango_id') IS NULL OR (datos->>'recibo_numero') IS NULL OR jsonb_typeof(datos->'formas') <> 'array' OR jsonb_array_length(datos->'formas') = 0 THEN
    RAISE EXCEPTION 'COBRO_INCOMPLETO';
  END IF;

  recibo_clave := (datos->>'recibo_uid')::uuid;
  recibo_numero := erp_v4.gs_asignar_numero(e,(datos->>'recibo_talonario_id')::bigint,(datos->>'usuario_id')::bigint,recibo_clave,terminal,(datos->>'recibo_rango_id')::bigint,(datos->>'recibo_numero')::bigint,ocurrido);

  INSERT INTO erp_v4.gf_recibo_cabecera(empresa_id,uid,tercero_id,apertura_id,numero_id,moneda_id,importe,fecha)
  VALUES (e, recibo_clave, (datos->>'cliente_id')::bigint, (datos->>'apertura_id')::bigint, recibo_numero, (datos->>'moneda_id')::bigint, (datos->>'total')::numeric, ocurrido)
  RETURNING id INTO recibo;

  FOR forma IN SELECT value FROM jsonb_array_elements(datos->'formas') LOOP
    INSERT INTO erp_v4.gf_recibo_forma(empresa_id,cabecera_id,medio_id,importe,referencia)
    VALUES (e, recibo, (forma->>'medio_id')::bigint, (forma->>'importe')::numeric, forma->>'referencia');
  END LOOP;

  SELECT COALESCE(SUM(importe),0) INTO total FROM erp_v4.gf_recibo_forma WHERE empresa_id=e AND cabecera_id=recibo;
  IF total <> (datos->>'total')::numeric THEN
    RAISE EXCEPTION 'COBRO_NO_CUADRA';
  END IF;
  PERFORM erp_v4.gf_confirmar(e, 'gf_recibo', recibo);
  RETURN venta;
END;
$$;

ALTER FUNCTION erp_v4.gp_integrar_venta_cobro(bigint,bigint,bigint,uuid,timestamptz,jsonb) SET search_path=erp_v4,pg_catalog;
REVOKE EXECUTE ON FUNCTION erp_v4.gp_integrar_venta_cobro(bigint,bigint,bigint,uuid,timestamptz,jsonb) FROM PUBLIC;
COMMIT;
