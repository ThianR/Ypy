-- GENESIS ERP V4 -- PostgreSQL 16+ -- instalacion NUEVA, no migracion.
-- Generado por herramientas/v4/generar.mjs; reglas en reglas.sql.
-- Ejecutar con psql -v ON_ERROR_STOP=1 -f creacionV4.sql BASE_NUEVA
-- No modifica V2 ni BIG_INT. Falla si erp_v4 ya existe, sin borrar datos.
-- Ver V4_DISENO_Y_REVISION.md: alcance, puesta en marcha y limites.
BEGIN;
CREATE SCHEMA erp_v4;
SET LOCAL search_path = erp_v4, pg_catalog;
CREATE DOMAIN cantidad AS NUMERIC(20,6);
CREATE DOMAIN importe AS NUMERIC(20,6);
CREATE DOMAIN factor AS NUMERIC(24,10) CHECK (VALUE > 0);
CREATE DOMAIN objeto_json AS JSONB CHECK (jsonb_typeof(VALUE) = 'object');

-- =====================================================================
-- 1. Organizacion, acceso y configuracion por empresa
-- =====================================================================

CREATE TABLE gs_moneda (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  codigo VARCHAR(3) NOT NULL UNIQUE CHECK (codigo ~ '^[A-Z]{3}$'), nombre TEXT NOT NULL,
  decimales SMALLINT NOT NULL CHECK (decimales BETWEEN 0 AND 6)
);

CREATE TABLE gs_empresa (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  codigo TEXT NOT NULL UNIQUE, razon_social TEXT NOT NULL, documento_fiscal TEXT NOT NULL,
  pais_codigo VARCHAR(2) NOT NULL, moneda_base_id BIGINT NOT NULL REFERENCES gs_moneda(id),
  zona_horaria TEXT NOT NULL DEFAULT 'America/Asuncion', activa BOOLEAN NOT NULL DEFAULT TRUE,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE gs_usuario (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  identificador UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(), login TEXT NOT NULL,
  password_hash TEXT, proveedor_identidad TEXT, sujeto_externo TEXT, nombre TEXT NOT NULL,
  activo BOOLEAN NOT NULL DEFAULT TRUE, creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (password_hash IS NOT NULL OR (proveedor_identidad IS NOT NULL AND sujeto_externo IS NOT NULL)),
  UNIQUE (proveedor_identidad, sujeto_externo)
);

CREATE UNIQUE INDEX uq_usuario_login ON gs_usuario(lower(login));

CREATE TABLE gs_usuario_empresa (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  usuario_id BIGINT NOT NULL REFERENCES gs_usuario(id), activo BOOLEAN NOT NULL DEFAULT TRUE,
  UNIQUE (empresa_id, usuario_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gs_sucursal (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  codigo TEXT NOT NULL, nombre TEXT NOT NULL, activa BOOLEAN NOT NULL DEFAULT TRUE,
  UNIQUE (empresa_id, codigo),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gs_rol (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  codigo TEXT NOT NULL, nombre TEXT NOT NULL, UNIQUE (empresa_id, codigo),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gs_permiso (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  codigo TEXT NOT NULL UNIQUE, descripcion TEXT NOT NULL
);

CREATE TABLE gs_rol_permiso (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  rol_id BIGINT NOT NULL, permiso_id BIGINT NOT NULL REFERENCES gs_permiso(id),
  FOREIGN KEY (empresa_id, rol_id) REFERENCES gs_rol(empresa_id, id) , UNIQUE (empresa_id, rol_id, permiso_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gs_usuario_rol (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  usuario_id BIGINT NOT NULL, rol_id BIGINT NOT NULL, sucursal_id BIGINT,
  FOREIGN KEY (empresa_id, usuario_id) REFERENCES gs_usuario_empresa(empresa_id, usuario_id),
  FOREIGN KEY (empresa_id, rol_id) REFERENCES gs_rol(empresa_id, id) , FOREIGN KEY (empresa_id, sucursal_id) REFERENCES gs_sucursal(empresa_id, id) ,
  UNIQUE NULLS NOT DISTINCT (empresa_id, usuario_id, rol_id, sucursal_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gs_modulo (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  codigo TEXT NOT NULL UNIQUE, descripcion TEXT NOT NULL
);

CREATE TABLE gs_empresa_modulo (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  modulo_id BIGINT NOT NULL REFERENCES gs_modulo(id), activo BOOLEAN NOT NULL DEFAULT TRUE,
  opciones objeto_json NOT NULL DEFAULT '{}', UNIQUE (empresa_id, modulo_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gs_cotizacion_moneda (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  moneda_origen_id BIGINT NOT NULL REFERENCES gs_moneda(id),
  moneda_destino_id BIGINT NOT NULL REFERENCES gs_moneda(id), fecha TIMESTAMPTZ NOT NULL,
  tipo TEXT NOT NULL CHECK (tipo IN ('COMPRA','VENTA','CONTABLE')), valor factor NOT NULL,
  CHECK (moneda_origen_id <> moneda_destino_id), UNIQUE (empresa_id, moneda_origen_id, moneda_destino_id, fecha, tipo),
  UNIQUE (empresa_id, id)
);

-- =====================================================================
-- 2. Terceros, contactos y condiciones
-- =====================================================================

CREATE TABLE gs_entidad (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  codigo TEXT NOT NULL, naturaleza TEXT NOT NULL CHECK (naturaleza IN ('PERSONA','ORGANIZACION')),
  nombre TEXT NOT NULL, activa BOOLEAN NOT NULL DEFAULT TRUE, atributos objeto_json NOT NULL DEFAULT '{}',
  UNIQUE (empresa_id, codigo),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gs_entidad_documento (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  entidad_id BIGINT NOT NULL, pais_codigo VARCHAR(2) NOT NULL,
  tipo TEXT NOT NULL, numero TEXT NOT NULL, principal BOOLEAN NOT NULL DEFAULT FALSE,
  FOREIGN KEY (empresa_id, entidad_id) REFERENCES gs_entidad(empresa_id, id) , UNIQUE (empresa_id, pais_codigo, tipo, numero),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gs_entidad_contacto (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  entidad_id BIGINT NOT NULL, tipo TEXT NOT NULL, valor TEXT NOT NULL,
  nombre_contacto TEXT, principal BOOLEAN NOT NULL DEFAULT FALSE, FOREIGN KEY (empresa_id, entidad_id) REFERENCES gs_entidad(empresa_id, id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gs_entidad_direccion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  entidad_id BIGINT NOT NULL, tipo TEXT NOT NULL, direccion TEXT NOT NULL,
  pais_codigo VARCHAR(2) NOT NULL, departamento TEXT, ciudad TEXT, barrio TEXT, codigo_postal TEXT,
  latitud NUMERIC(10,7) CHECK (latitud BETWEEN -90 AND 90), longitud NUMERIC(10,7) CHECK (longitud BETWEEN -180 AND 180),
  principal BOOLEAN NOT NULL DEFAULT FALSE, FOREIGN KEY (empresa_id, entidad_id) REFERENCES gs_entidad(empresa_id, id),
  UNIQUE (empresa_id, id)
);

CREATE UNIQUE INDEX uq_gs_entidad_documento_principal ON gs_entidad_documento(empresa_id, entidad_id) WHERE principal;

CREATE UNIQUE INDEX uq_gs_entidad_contacto_principal ON gs_entidad_contacto(empresa_id, entidad_id, tipo) WHERE principal;

CREATE UNIQUE INDEX uq_gs_entidad_direccion_principal ON gs_entidad_direccion(empresa_id, entidad_id, tipo) WHERE principal;

CREATE TABLE gf_condicion_pago (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  codigo TEXT NOT NULL, descripcion TEXT NOT NULL, activa BOOLEAN NOT NULL DEFAULT TRUE,
  UNIQUE (empresa_id, codigo),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gf_condicion_pago_cuota (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  condicion_id BIGINT NOT NULL, numero INTEGER NOT NULL CHECK (numero > 0),
  dias_desde_emision INTEGER NOT NULL CHECK (dias_desde_emision >= 0), porcentaje NUMERIC(9,6) NOT NULL CHECK (porcentaje > 0 AND porcentaje <= 100),
  FOREIGN KEY (empresa_id, condicion_id) REFERENCES gf_condicion_pago(empresa_id, id) , UNIQUE (empresa_id, condicion_id, numero),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_cliente (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  entidad_id BIGINT NOT NULL, condicion_pago_id BIGINT, moneda_id BIGINT NOT NULL REFERENCES gs_moneda(id),
  limite_credito importe NOT NULL DEFAULT 0 CHECK (limite_credito >= 0), activo BOOLEAN NOT NULL DEFAULT TRUE, FOREIGN KEY (empresa_id, entidad_id) REFERENCES gs_entidad(empresa_id, id) , FOREIGN KEY (empresa_id, condicion_pago_id) REFERENCES gf_condicion_pago(empresa_id, id) , UNIQUE (empresa_id, entidad_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gc_proveedor (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  entidad_id BIGINT NOT NULL, condicion_pago_id BIGINT, moneda_id BIGINT NOT NULL REFERENCES gs_moneda(id),
  codigo_cliente_proveedor TEXT, activo BOOLEAN NOT NULL DEFAULT TRUE, FOREIGN KEY (empresa_id, entidad_id) REFERENCES gs_entidad(empresa_id, id) , FOREIGN KEY (empresa_id, condicion_pago_id) REFERENCES gf_condicion_pago(empresa_id, id) , UNIQUE (empresa_id, entidad_id),
  UNIQUE (empresa_id, id)
);

-- =====================================================================
-- 3. Catalogo, presentaciones, trazabilidad, impuestos y precios
-- =====================================================================

CREATE TABLE gi_unidad_medida (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  codigo TEXT NOT NULL UNIQUE, nombre TEXT NOT NULL, dimension TEXT NOT NULL,
  admite_fraccion BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE gi_categoria (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  codigo TEXT NOT NULL, nombre TEXT NOT NULL, padre_id BIGINT,
  FOREIGN KEY (empresa_id, padre_id) REFERENCES gi_categoria(empresa_id, id) , CHECK (padre_id IS NULL OR padre_id <> id), UNIQUE (empresa_id, codigo),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gi_producto (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  codigo TEXT NOT NULL, nombre TEXT NOT NULL, categoria_id BIGINT,
  tipo TEXT NOT NULL CHECK (tipo IN ('MERCADERIA','SERVICIO','KIT')), atributos objeto_json NOT NULL DEFAULT '{}',
  FOREIGN KEY (empresa_id, categoria_id) REFERENCES gi_categoria(empresa_id, id) , UNIQUE (empresa_id, codigo),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gi_articulo (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  producto_id BIGINT NOT NULL, codigo TEXT NOT NULL, descripcion TEXT NOT NULL,
  unidad_base_id BIGINT NOT NULL REFERENCES gi_unidad_medida(id), mueve_stock BOOLEAN NOT NULL DEFAULT TRUE,
  controla_lote BOOLEAN NOT NULL DEFAULT FALSE, controla_serie BOOLEAN NOT NULL DEFAULT FALSE,
  exige_vencimiento BOOLEAN NOT NULL DEFAULT FALSE, vendible BOOLEAN NOT NULL DEFAULT TRUE,
  comprable BOOLEAN NOT NULL DEFAULT TRUE, peso_kg cantidad CHECK (peso_kg >= 0),
  atributos objeto_json NOT NULL DEFAULT '{}', activo BOOLEAN NOT NULL DEFAULT TRUE,
  FOREIGN KEY (empresa_id, producto_id) REFERENCES gi_producto(empresa_id, id) , CHECK (NOT exige_vencimiento OR controla_lote),
  CHECK (mueve_stock OR NOT (controla_lote OR controla_serie)), UNIQUE (empresa_id, codigo),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gi_presentacion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  articulo_id BIGINT NOT NULL, codigo TEXT NOT NULL,
  unidad_id BIGINT NOT NULL REFERENCES gi_unidad_medida(id), factor_base factor NOT NULL,
  permite_fraccion BOOLEAN NOT NULL DEFAULT FALSE, activa BOOLEAN NOT NULL DEFAULT TRUE,
  FOREIGN KEY (empresa_id, articulo_id) REFERENCES gi_articulo(empresa_id, id) , UNIQUE (empresa_id, articulo_id, codigo), UNIQUE (empresa_id, id, articulo_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gi_codigo_barra (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  presentacion_id BIGINT NOT NULL, codigo TEXT NOT NULL,
  FOREIGN KEY (empresa_id, presentacion_id) REFERENCES gi_presentacion(empresa_id, id) , UNIQUE (empresa_id, codigo),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gp_regla_balanza (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  prefijo TEXT NOT NULL, longitud SMALLINT NOT NULL CHECK (longitud > 0),
  inicio_producto SMALLINT NOT NULL CHECK (inicio_producto > 0), largo_producto SMALLINT NOT NULL CHECK (largo_producto > 0),
  inicio_valor SMALLINT NOT NULL CHECK (inicio_valor > 0), largo_valor SMALLINT NOT NULL CHECK (largo_valor > 0),
  decimales SMALLINT NOT NULL CHECK (decimales BETWEEN 0 AND 6), contenido TEXT NOT NULL CHECK (contenido IN ('PESO','PRECIO')),
  CHECK (inicio_producto + largo_producto - 1 <= longitud), CHECK (inicio_valor + largo_valor - 1 <= longitud), UNIQUE (empresa_id, prefijo),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gi_kit_componente (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  articulo_padre_id BIGINT NOT NULL, articulo_hijo_id BIGINT NOT NULL,
  cantidad_base cantidad NOT NULL CHECK (cantidad_base > 0),
  FOREIGN KEY (empresa_id, articulo_padre_id) REFERENCES gi_articulo(empresa_id, id) , FOREIGN KEY (empresa_id, articulo_hijo_id) REFERENCES gi_articulo(empresa_id, id) ,
  CHECK (articulo_padre_id <> articulo_hijo_id), UNIQUE (empresa_id, articulo_padre_id, articulo_hijo_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gi_lote (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  articulo_id BIGINT NOT NULL, codigo TEXT NOT NULL, fecha_fabricacion DATE, fecha_vencimiento DATE,
  FOREIGN KEY (empresa_id, articulo_id) REFERENCES gi_articulo(empresa_id, id) , CHECK (fecha_vencimiento >= fecha_fabricacion),
  UNIQUE (empresa_id, articulo_id, codigo), UNIQUE (empresa_id, id, articulo_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gi_serie (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  articulo_id BIGINT NOT NULL, numero TEXT NOT NULL,
  FOREIGN KEY (empresa_id, articulo_id) REFERENCES gi_articulo(empresa_id, id) , UNIQUE (empresa_id, articulo_id, numero), UNIQUE (empresa_id, id, articulo_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gs_impuesto (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  codigo TEXT NOT NULL, nombre TEXT NOT NULL, jurisdiccion TEXT NOT NULL,
  UNIQUE (empresa_id, codigo),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gs_impuesto_tasa (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  impuesto_id BIGINT NOT NULL, desde DATE NOT NULL, hasta DATE,
  porcentaje NUMERIC(9,6) NOT NULL CHECK (porcentaje BETWEEN 0 AND 100), codigo_fiscal TEXT,
  FOREIGN KEY (empresa_id, impuesto_id) REFERENCES gs_impuesto(empresa_id, id) , CHECK (hasta IS NULL OR hasta >= desde), UNIQUE (empresa_id, impuesto_id, desde),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gi_articulo_impuesto (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  articulo_id BIGINT NOT NULL, impuesto_id BIGINT NOT NULL,
  operacion TEXT NOT NULL CHECK (operacion IN ('COMPRA','VENTA')), FOREIGN KEY (empresa_id, articulo_id) REFERENCES gi_articulo(empresa_id, id) ,
  FOREIGN KEY (empresa_id, impuesto_id) REFERENCES gs_impuesto(empresa_id, id) , UNIQUE (empresa_id, articulo_id, impuesto_id, operacion),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_lista_precio (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  codigo TEXT NOT NULL, nombre TEXT NOT NULL, moneda_id BIGINT NOT NULL REFERENCES gs_moneda(id),
  incluye_impuestos BOOLEAN NOT NULL, UNIQUE (empresa_id, codigo),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_precio (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  lista_id BIGINT NOT NULL, presentacion_id BIGINT NOT NULL, desde TIMESTAMPTZ NOT NULL,
  hasta TIMESTAMPTZ, precio importe NOT NULL CHECK (precio >= 0), minimo cantidad NOT NULL DEFAULT 1 CHECK (minimo > 0),
  FOREIGN KEY (empresa_id, lista_id) REFERENCES gv_lista_precio(empresa_id, id) , FOREIGN KEY (empresa_id, presentacion_id) REFERENCES gi_presentacion(empresa_id, id) ,
  CHECK (hasta IS NULL OR hasta > desde), UNIQUE (empresa_id, lista_id, presentacion_id, desde, minimo),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_promocion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  codigo TEXT NOT NULL, nombre TEXT NOT NULL, desde TIMESTAMPTZ NOT NULL, hasta TIMESTAMPTZ NOT NULL,
  prioridad INTEGER NOT NULL DEFAULT 0, acumulable BOOLEAN NOT NULL DEFAULT FALSE,
  tipo TEXT NOT NULL CHECK (tipo IN ('DESCUENTO_PORCENTAJE','PRECIO_FIJO','N_POR_M')),
  parametros objeto_json NOT NULL, activa BOOLEAN NOT NULL DEFAULT TRUE,
  CHECK (hasta > desde), UNIQUE (empresa_id, codigo),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_promocion_articulo (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  promocion_id BIGINT NOT NULL, articulo_id BIGINT NOT NULL,
  FOREIGN KEY (empresa_id, promocion_id) REFERENCES gv_promocion(empresa_id, id) , FOREIGN KEY (empresa_id, articulo_id) REFERENCES gi_articulo(empresa_id, id) , UNIQUE (empresa_id, promocion_id, articulo_id),
  UNIQUE (empresa_id, id)
);

-- =====================================================================
-- 4. Acceso, modalidad fiscal y numeracion general
-- =====================================================================

CREATE TABLE gs_usuario_sucursal (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  usuario_id BIGINT NOT NULL, sucursal_id BIGINT NOT NULL,
  puede_emitir BOOLEAN NOT NULL DEFAULT FALSE, supervisor BOOLEAN NOT NULL DEFAULT FALSE, activo BOOLEAN NOT NULL DEFAULT TRUE,
  FOREIGN KEY (empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id),
  FOREIGN KEY (empresa_id, sucursal_id) REFERENCES gs_sucursal(empresa_id, id) , UNIQUE(empresa_id,usuario_id,sucursal_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gp_terminal (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  sucursal_id BIGINT NOT NULL, codigo TEXT NOT NULL, uid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
  activa BOOLEAN NOT NULL DEFAULT TRUE, horas_permiso_offline INTEGER NOT NULL DEFAULT 72 CHECK(horas_permiso_offline BETWEEN 1 AND 720),
  ultima_sincronizacion TIMESTAMPTZ, FOREIGN KEY (empresa_id, sucursal_id) REFERENCES gs_sucursal(empresa_id, id) , UNIQUE(empresa_id,codigo), UNIQUE(empresa_id,id,sucursal_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gs_modalidad_emision (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  codigo TEXT NOT NULL UNIQUE, nombre TEXT NOT NULL, fiscal BOOLEAN NOT NULL,
  integracion TEXT NOT NULL CHECK(integracion IN ('INTERNA','REGISTRO','SIFEN','EXTERNA'))
);

CREATE TABLE gs_habilitacion_fiscal (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  modalidad_id BIGINT NOT NULL REFERENCES gs_modalidad_emision(id),
  desde DATE NOT NULL, hasta DATE, autorizacion TEXT NOT NULL, fundamento TEXT NOT NULL,
  estado TEXT NOT NULL CHECK(estado IN ('PENDIENTE','VALIDADA','REVOCADA')), evidencia TEXT,
  CHECK(hasta IS NULL OR hasta>=desde), UNIQUE(empresa_id,id,modalidad_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gs_clase_documento (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  codigo TEXT NOT NULL UNIQUE, modulo TEXT NOT NULL CHECK(modulo IN ('gv','gc','gi','gf','gl')),
  descripcion TEXT NOT NULL
);

CREATE TABLE gs_tipo_comprobante (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  codigo TEXT NOT NULL, clase_id BIGINT NOT NULL REFERENCES gs_clase_documento(id),
  efecto_stock_predeterminado TEXT NOT NULL CHECK(efecto_stock_predeterminado IN ('ENTRADA','SALIDA','NINGUNO')),
  efecto_financiero SMALLINT NOT NULL DEFAULT 1 CHECK(efecto_financiero IN (-1,0,1)),
  permite_excepcion BOOLEAN NOT NULL DEFAULT FALSE, activo BOOLEAN NOT NULL DEFAULT TRUE, UNIQUE(empresa_id,codigo),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gs_talonario (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  sucursal_id BIGINT NOT NULL, clase_id BIGINT NOT NULL REFERENCES gs_clase_documento(id),
  modalidad_id BIGINT NOT NULL REFERENCES gs_modalidad_emision(id), habilitacion_id BIGINT,
  codigo TEXT NOT NULL, serie TEXT NOT NULL DEFAULT '', establecimiento VARCHAR(3), punto_expedicion VARCHAR(3), timbrado TEXT,
  desde DATE NOT NULL, hasta DATE, inicial BIGINT NOT NULL CHECK(inicial>0), final BIGINT NOT NULL,
  siguiente BIGINT NOT NULL, publico_sucursal BOOLEAN NOT NULL DEFAULT FALSE, activo BOOLEAN NOT NULL DEFAULT TRUE,
  FOREIGN KEY (empresa_id, sucursal_id) REFERENCES gs_sucursal(empresa_id, id) , FOREIGN KEY(empresa_id,habilitacion_id,modalidad_id) REFERENCES gs_habilitacion_fiscal(empresa_id,id,modalidad_id),
  CHECK(final>=inicial AND siguiente BETWEEN inicial AND final+1), CHECK(hasta IS NULL OR hasta>=desde),
  CHECK(establecimiento IS NULL OR establecimiento ~ '^[0-9]{3}$'), CHECK(punto_expedicion IS NULL OR punto_expedicion ~ '^[0-9]{3}$'),
  UNIQUE(empresa_id,codigo), UNIQUE NULLS NOT DISTINCT(empresa_id,clase_id,modalidad_id,timbrado,establecimiento,punto_expedicion,serie),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gs_talonario_usuario (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  talonario_id BIGINT NOT NULL, usuario_id BIGINT NOT NULL,
  FOREIGN KEY (empresa_id, talonario_id) REFERENCES gs_talonario(empresa_id, id) , FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id), UNIQUE(empresa_id,talonario_id,usuario_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gs_talonario_alternativo (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  principal_id BIGINT NOT NULL, alternativo_id BIGINT NOT NULL, prioridad INTEGER NOT NULL CHECK(prioridad>0),
  FOREIGN KEY (empresa_id, principal_id) REFERENCES gs_talonario(empresa_id, id) , FOREIGN KEY (empresa_id, alternativo_id) REFERENCES gs_talonario(empresa_id, id) , CHECK(principal_id<>alternativo_id), UNIQUE(empresa_id,principal_id,prioridad),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gs_talonario_rango_terminal (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  talonario_id BIGINT NOT NULL, terminal_id BIGINT NOT NULL,
  uid UUID NOT NULL UNIQUE, inicio BIGINT NOT NULL, fin BIGINT NOT NULL,
  reserva_adicional BOOLEAN NOT NULL DEFAULT FALSE, asignado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (empresa_id, talonario_id) REFERENCES gs_talonario(empresa_id, id) , FOREIGN KEY (empresa_id, terminal_id) REFERENCES gp_terminal(empresa_id, id) , CHECK(inicio>0 AND fin>=inicio), UNIQUE(empresa_id,id,talonario_id,terminal_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gs_talonario_numero (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  talonario_id BIGINT NOT NULL, numero BIGINT NOT NULL CHECK(numero>0), uid UUID NOT NULL UNIQUE,
  usuario_id BIGINT NOT NULL, terminal_id BIGINT, rango_id BIGINT, clase_id BIGINT NOT NULL REFERENCES gs_clase_documento(id),
  emitido_en TIMESTAMPTZ NOT NULL DEFAULT now(), registrado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  estado TEXT NOT NULL DEFAULT 'ASIGNADO' CHECK(estado IN ('ASIGNADO','ANULADO')),
  FOREIGN KEY (empresa_id, talonario_id) REFERENCES gs_talonario(empresa_id, id) , FOREIGN KEY (empresa_id, terminal_id) REFERENCES gp_terminal(empresa_id, id) ,
  FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id),
  FOREIGN KEY(empresa_id,rango_id,talonario_id,terminal_id) REFERENCES gs_talonario_rango_terminal(empresa_id,id,talonario_id,terminal_id),
  CHECK(rango_id IS NULL OR terminal_id IS NOT NULL), UNIQUE(empresa_id,talonario_id,numero),
  UNIQUE (empresa_id, id)
);

-- =====================================================================
-- 5. Posiciones fisicas y valoracion por empresa/articulo
-- =====================================================================

ALTER TABLE gi_articulo ADD COLUMN costo_referencia importe CHECK(costo_referencia>=0);

CREATE TABLE gi_deposito (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  sucursal_id BIGINT NOT NULL, codigo TEXT NOT NULL, nombre TEXT NOT NULL,
 FOREIGN KEY (empresa_id, sucursal_id) REFERENCES gs_sucursal(empresa_id, id) , UNIQUE(empresa_id,codigo),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gi_ubicacion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  deposito_id BIGINT NOT NULL, codigo TEXT NOT NULL, FOREIGN KEY (empresa_id, deposito_id) REFERENCES gi_deposito(empresa_id, id) , UNIQUE(empresa_id,deposito_id,codigo),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gi_posicion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  ubicacion_id BIGINT NOT NULL, articulo_id BIGINT NOT NULL, lote_id BIGINT, serie_id BIGINT,
 FOREIGN KEY (empresa_id, ubicacion_id) REFERENCES gi_ubicacion(empresa_id, id) , FOREIGN KEY (empresa_id, articulo_id) REFERENCES gi_articulo(empresa_id, id) ,
 FOREIGN KEY(empresa_id,lote_id,articulo_id) REFERENCES gi_lote(empresa_id,id,articulo_id),
 FOREIGN KEY(empresa_id,serie_id,articulo_id) REFERENCES gi_serie(empresa_id,id,articulo_id),
 UNIQUE NULLS NOT DISTINCT(empresa_id,ubicacion_id,articulo_id,lote_id,serie_id), UNIQUE(empresa_id,id,articulo_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gi_existencia (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  posicion_id BIGINT NOT NULL, cantidad cantidad NOT NULL DEFAULT 0, reservada cantidad NOT NULL DEFAULT 0 CHECK(reservada>=0),
 disponible cantidad GENERATED ALWAYS AS(cantidad-reservada) STORED,
 FOREIGN KEY (empresa_id, posicion_id) REFERENCES gi_posicion(empresa_id, id) , UNIQUE(empresa_id,posicion_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gi_valoracion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  articulo_id BIGINT NOT NULL, cantidad cantidad NOT NULL DEFAULT 0,
 valor NUMERIC(28,10) NOT NULL DEFAULT 0, ultimo_costo NUMERIC(28,10) CHECK(ultimo_costo>=0),
 pendiente BOOLEAN NOT NULL DEFAULT FALSE, version BIGINT NOT NULL DEFAULT 0,
 FOREIGN KEY (empresa_id, articulo_id) REFERENCES gi_articulo(empresa_id, id) , UNIQUE(empresa_id,articulo_id),
  UNIQUE (empresa_id, id)
);

-- =====================================================================
-- 6. Cabeceras y detalles propios por modulo
-- =====================================================================

CREATE TABLE gc_orden_compra_cabecera (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  uid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE, sucursal_id BIGINT NOT NULL,
  usuario_id BIGINT NOT NULL, tipo_id BIGINT NOT NULL, numero_id BIGINT UNIQUE, terminal_id BIGINT,
  tercero_id BIGINT NOT NULL, FOREIGN KEY (empresa_id, tercero_id) REFERENCES gc_proveedor(empresa_id, id) ,
  moneda_id BIGINT NOT NULL REFERENCES gs_moneda(id), cambio_base factor NOT NULL DEFAULT 1,
  fecha_operacion TIMESTAMPTZ NOT NULL DEFAULT now(), confirmado_en TIMESTAMPTZ,
  estado TEXT NOT NULL DEFAULT 'BORRADOR' CHECK(estado IN ('BORRADOR','CONFIRMADO','REVERTIDO')),
  efecto_stock_aplicado TEXT CHECK(efecto_stock_aplicado IN ('ENTRADA','SALIDA','NINGUNO')),
  motivo_excepcion TEXT, total importe NOT NULL DEFAULT 0 CHECK(total>=0), tercero_snapshot objeto_json NOT NULL DEFAULT '{}',
  
  
  
  FOREIGN KEY (empresa_id, sucursal_id) REFERENCES gs_sucursal(empresa_id, id) , FOREIGN KEY (empresa_id, tipo_id) REFERENCES gs_tipo_comprobante(empresa_id, id) , FOREIGN KEY (empresa_id, numero_id) REFERENCES gs_talonario_numero(empresa_id, id) , FOREIGN KEY (empresa_id, terminal_id) REFERENCES gp_terminal(empresa_id, id) ,
  FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gc_orden_compra_detalle (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, renglon INTEGER NOT NULL CHECK(renglon>0),
  articulo_id BIGINT NOT NULL, presentacion_id BIGINT NOT NULL, descripcion TEXT NOT NULL,
  cantidad cantidad NOT NULL CHECK(cantidad>0), factor_base factor NOT NULL,
  cantidad_base cantidad GENERATED ALWAYS AS(cantidad*factor_base) STORED,
  precio importe NOT NULL CHECK(precio>=0), descuento importe NOT NULL DEFAULT 0 CHECK(descuento>=0 AND descuento<=cantidad*precio),
  tasa_impuesto NUMERIC(9,6) NOT NULL DEFAULT 0 CHECK(tasa_impuesto BETWEEN 0 AND 100),
  base_imponible importe NOT NULL DEFAULT 0 CHECK(base_imponible>=0), impuesto importe NOT NULL DEFAULT 0 CHECK(impuesto>=0),
  importe_total importe GENERATED ALWAYS AS(cantidad*precio-descuento+impuesto) STORED,
  costo_unitario_base NUMERIC(28,10) CHECK(costo_unitario_base>=0), posicion_id BIGINT, destino_id BIGINT,
  FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gc_orden_compra_cabecera(empresa_id, id) , FOREIGN KEY (empresa_id, articulo_id) REFERENCES gi_articulo(empresa_id, id) ,
  FOREIGN KEY(empresa_id,presentacion_id,articulo_id) REFERENCES gi_presentacion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,posicion_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,destino_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  UNIQUE(empresa_id,cabecera_id,renglon), UNIQUE(empresa_id,id,articulo_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gc_recepcion_cabecera (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  uid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE, sucursal_id BIGINT NOT NULL,
  usuario_id BIGINT NOT NULL, tipo_id BIGINT NOT NULL, numero_id BIGINT UNIQUE, terminal_id BIGINT,
  tercero_id BIGINT NOT NULL, FOREIGN KEY (empresa_id, tercero_id) REFERENCES gc_proveedor(empresa_id, id) ,
  moneda_id BIGINT NOT NULL REFERENCES gs_moneda(id), cambio_base factor NOT NULL DEFAULT 1,
  fecha_operacion TIMESTAMPTZ NOT NULL DEFAULT now(), confirmado_en TIMESTAMPTZ,
  estado TEXT NOT NULL DEFAULT 'BORRADOR' CHECK(estado IN ('BORRADOR','CONFIRMADO','REVERTIDO')),
  efecto_stock_aplicado TEXT CHECK(efecto_stock_aplicado IN ('ENTRADA','SALIDA','NINGUNO')),
  motivo_excepcion TEXT, total importe NOT NULL DEFAULT 0 CHECK(total>=0), tercero_snapshot objeto_json NOT NULL DEFAULT '{}',
  
  
  
  FOREIGN KEY (empresa_id, sucursal_id) REFERENCES gs_sucursal(empresa_id, id) , FOREIGN KEY (empresa_id, tipo_id) REFERENCES gs_tipo_comprobante(empresa_id, id) , FOREIGN KEY (empresa_id, numero_id) REFERENCES gs_talonario_numero(empresa_id, id) , FOREIGN KEY (empresa_id, terminal_id) REFERENCES gp_terminal(empresa_id, id) ,
  FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gc_recepcion_detalle (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, renglon INTEGER NOT NULL CHECK(renglon>0),
  articulo_id BIGINT NOT NULL, presentacion_id BIGINT NOT NULL, descripcion TEXT NOT NULL,
  cantidad cantidad NOT NULL CHECK(cantidad>0), factor_base factor NOT NULL,
  cantidad_base cantidad GENERATED ALWAYS AS(cantidad*factor_base) STORED,
  precio importe NOT NULL CHECK(precio>=0), descuento importe NOT NULL DEFAULT 0 CHECK(descuento>=0 AND descuento<=cantidad*precio),
  tasa_impuesto NUMERIC(9,6) NOT NULL DEFAULT 0 CHECK(tasa_impuesto BETWEEN 0 AND 100),
  base_imponible importe NOT NULL DEFAULT 0 CHECK(base_imponible>=0), impuesto importe NOT NULL DEFAULT 0 CHECK(impuesto>=0),
  importe_total importe GENERATED ALWAYS AS(cantidad*precio-descuento+impuesto) STORED,
  costo_unitario_base NUMERIC(28,10) CHECK(costo_unitario_base>=0), posicion_id BIGINT, destino_id BIGINT,
  FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gc_recepcion_cabecera(empresa_id, id) , FOREIGN KEY (empresa_id, articulo_id) REFERENCES gi_articulo(empresa_id, id) ,
  FOREIGN KEY(empresa_id,presentacion_id,articulo_id) REFERENCES gi_presentacion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,posicion_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,destino_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  UNIQUE(empresa_id,cabecera_id,renglon), UNIQUE(empresa_id,id,articulo_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gc_comprobante_cabecera (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  uid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE, sucursal_id BIGINT NOT NULL,
  usuario_id BIGINT NOT NULL, tipo_id BIGINT NOT NULL, numero_id BIGINT UNIQUE, terminal_id BIGINT,
  tercero_id BIGINT NOT NULL, FOREIGN KEY (empresa_id, tercero_id) REFERENCES gc_proveedor(empresa_id, id) ,
  moneda_id BIGINT NOT NULL REFERENCES gs_moneda(id), cambio_base factor NOT NULL DEFAULT 1,
  fecha_operacion TIMESTAMPTZ NOT NULL DEFAULT now(), confirmado_en TIMESTAMPTZ,
  estado TEXT NOT NULL DEFAULT 'BORRADOR' CHECK(estado IN ('BORRADOR','CONFIRMADO','REVERTIDO')),
  efecto_stock_aplicado TEXT CHECK(efecto_stock_aplicado IN ('ENTRADA','SALIDA','NINGUNO')),
  motivo_excepcion TEXT, total importe NOT NULL DEFAULT 0 CHECK(total>=0), tercero_snapshot objeto_json NOT NULL DEFAULT '{}',
  numero_proveedor TEXT NOT NULL, timbrado_proveedor TEXT NOT NULL DEFAULT '', UNIQUE(empresa_id,tercero_id,tipo_id,timbrado_proveedor,numero_proveedor),
  
  
  FOREIGN KEY (empresa_id, sucursal_id) REFERENCES gs_sucursal(empresa_id, id) , FOREIGN KEY (empresa_id, tipo_id) REFERENCES gs_tipo_comprobante(empresa_id, id) , FOREIGN KEY (empresa_id, numero_id) REFERENCES gs_talonario_numero(empresa_id, id) , FOREIGN KEY (empresa_id, terminal_id) REFERENCES gp_terminal(empresa_id, id) ,
  FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gc_comprobante_detalle (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, renglon INTEGER NOT NULL CHECK(renglon>0),
  articulo_id BIGINT NOT NULL, presentacion_id BIGINT NOT NULL, descripcion TEXT NOT NULL,
  cantidad cantidad NOT NULL CHECK(cantidad>0), factor_base factor NOT NULL,
  cantidad_base cantidad GENERATED ALWAYS AS(cantidad*factor_base) STORED,
  precio importe NOT NULL CHECK(precio>=0), descuento importe NOT NULL DEFAULT 0 CHECK(descuento>=0 AND descuento<=cantidad*precio),
  tasa_impuesto NUMERIC(9,6) NOT NULL DEFAULT 0 CHECK(tasa_impuesto BETWEEN 0 AND 100),
  base_imponible importe NOT NULL DEFAULT 0 CHECK(base_imponible>=0), impuesto importe NOT NULL DEFAULT 0 CHECK(impuesto>=0),
  importe_total importe GENERATED ALWAYS AS(cantidad*precio-descuento+impuesto) STORED,
  costo_unitario_base NUMERIC(28,10) CHECK(costo_unitario_base>=0), posicion_id BIGINT, destino_id BIGINT,
  FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gc_comprobante_cabecera(empresa_id, id) , FOREIGN KEY (empresa_id, articulo_id) REFERENCES gi_articulo(empresa_id, id) ,
  FOREIGN KEY(empresa_id,presentacion_id,articulo_id) REFERENCES gi_presentacion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,posicion_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,destino_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  UNIQUE(empresa_id,cabecera_id,renglon), UNIQUE(empresa_id,id,articulo_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_cotizacion_cabecera (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  uid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE, sucursal_id BIGINT NOT NULL,
  usuario_id BIGINT NOT NULL, tipo_id BIGINT NOT NULL, numero_id BIGINT UNIQUE, terminal_id BIGINT,
  tercero_id BIGINT NOT NULL, FOREIGN KEY (empresa_id, tercero_id) REFERENCES gv_cliente(empresa_id, id) ,
  moneda_id BIGINT NOT NULL REFERENCES gs_moneda(id), cambio_base factor NOT NULL DEFAULT 1,
  fecha_operacion TIMESTAMPTZ NOT NULL DEFAULT now(), confirmado_en TIMESTAMPTZ,
  estado TEXT NOT NULL DEFAULT 'BORRADOR' CHECK(estado IN ('BORRADOR','CONFIRMADO','REVERTIDO')),
  efecto_stock_aplicado TEXT CHECK(efecto_stock_aplicado IN ('ENTRADA','SALIDA','NINGUNO')),
  motivo_excepcion TEXT, total importe NOT NULL DEFAULT 0 CHECK(total>=0), tercero_snapshot objeto_json NOT NULL DEFAULT '{}',
  
  
  
  FOREIGN KEY (empresa_id, sucursal_id) REFERENCES gs_sucursal(empresa_id, id) , FOREIGN KEY (empresa_id, tipo_id) REFERENCES gs_tipo_comprobante(empresa_id, id) , FOREIGN KEY (empresa_id, numero_id) REFERENCES gs_talonario_numero(empresa_id, id) , FOREIGN KEY (empresa_id, terminal_id) REFERENCES gp_terminal(empresa_id, id) ,
  FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_cotizacion_detalle (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, renglon INTEGER NOT NULL CHECK(renglon>0),
  articulo_id BIGINT NOT NULL, presentacion_id BIGINT NOT NULL, descripcion TEXT NOT NULL,
  cantidad cantidad NOT NULL CHECK(cantidad>0), factor_base factor NOT NULL,
  cantidad_base cantidad GENERATED ALWAYS AS(cantidad*factor_base) STORED,
  precio importe NOT NULL CHECK(precio>=0), descuento importe NOT NULL DEFAULT 0 CHECK(descuento>=0 AND descuento<=cantidad*precio),
  tasa_impuesto NUMERIC(9,6) NOT NULL DEFAULT 0 CHECK(tasa_impuesto BETWEEN 0 AND 100),
  base_imponible importe NOT NULL DEFAULT 0 CHECK(base_imponible>=0), impuesto importe NOT NULL DEFAULT 0 CHECK(impuesto>=0),
  importe_total importe GENERATED ALWAYS AS(cantidad*precio-descuento+impuesto) STORED,
  costo_unitario_base NUMERIC(28,10) CHECK(costo_unitario_base>=0), posicion_id BIGINT, destino_id BIGINT,
  FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gv_cotizacion_cabecera(empresa_id, id) , FOREIGN KEY (empresa_id, articulo_id) REFERENCES gi_articulo(empresa_id, id) ,
  FOREIGN KEY(empresa_id,presentacion_id,articulo_id) REFERENCES gi_presentacion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,posicion_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,destino_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  UNIQUE(empresa_id,cabecera_id,renglon), UNIQUE(empresa_id,id,articulo_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_pedido_cabecera (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  uid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE, sucursal_id BIGINT NOT NULL,
  usuario_id BIGINT NOT NULL, tipo_id BIGINT NOT NULL, numero_id BIGINT UNIQUE, terminal_id BIGINT,
  tercero_id BIGINT NOT NULL, FOREIGN KEY (empresa_id, tercero_id) REFERENCES gv_cliente(empresa_id, id) ,
  moneda_id BIGINT NOT NULL REFERENCES gs_moneda(id), cambio_base factor NOT NULL DEFAULT 1,
  fecha_operacion TIMESTAMPTZ NOT NULL DEFAULT now(), confirmado_en TIMESTAMPTZ,
  estado TEXT NOT NULL DEFAULT 'BORRADOR' CHECK(estado IN ('BORRADOR','CONFIRMADO','REVERTIDO')),
  efecto_stock_aplicado TEXT CHECK(efecto_stock_aplicado IN ('ENTRADA','SALIDA','NINGUNO')),
  motivo_excepcion TEXT, total importe NOT NULL DEFAULT 0 CHECK(total>=0), tercero_snapshot objeto_json NOT NULL DEFAULT '{}',
  
  
  
  FOREIGN KEY (empresa_id, sucursal_id) REFERENCES gs_sucursal(empresa_id, id) , FOREIGN KEY (empresa_id, tipo_id) REFERENCES gs_tipo_comprobante(empresa_id, id) , FOREIGN KEY (empresa_id, numero_id) REFERENCES gs_talonario_numero(empresa_id, id) , FOREIGN KEY (empresa_id, terminal_id) REFERENCES gp_terminal(empresa_id, id) ,
  FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_pedido_detalle (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, renglon INTEGER NOT NULL CHECK(renglon>0),
  articulo_id BIGINT NOT NULL, presentacion_id BIGINT NOT NULL, descripcion TEXT NOT NULL,
  cantidad cantidad NOT NULL CHECK(cantidad>0), factor_base factor NOT NULL,
  cantidad_base cantidad GENERATED ALWAYS AS(cantidad*factor_base) STORED,
  precio importe NOT NULL CHECK(precio>=0), descuento importe NOT NULL DEFAULT 0 CHECK(descuento>=0 AND descuento<=cantidad*precio),
  tasa_impuesto NUMERIC(9,6) NOT NULL DEFAULT 0 CHECK(tasa_impuesto BETWEEN 0 AND 100),
  base_imponible importe NOT NULL DEFAULT 0 CHECK(base_imponible>=0), impuesto importe NOT NULL DEFAULT 0 CHECK(impuesto>=0),
  importe_total importe GENERATED ALWAYS AS(cantidad*precio-descuento+impuesto) STORED,
  costo_unitario_base NUMERIC(28,10) CHECK(costo_unitario_base>=0), posicion_id BIGINT, destino_id BIGINT,
  FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gv_pedido_cabecera(empresa_id, id) , FOREIGN KEY (empresa_id, articulo_id) REFERENCES gi_articulo(empresa_id, id) ,
  FOREIGN KEY(empresa_id,presentacion_id,articulo_id) REFERENCES gi_presentacion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,posicion_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,destino_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  UNIQUE(empresa_id,cabecera_id,renglon), UNIQUE(empresa_id,id,articulo_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_comprobante_cabecera (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  uid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE, sucursal_id BIGINT NOT NULL,
  usuario_id BIGINT NOT NULL, tipo_id BIGINT NOT NULL, numero_id BIGINT UNIQUE, terminal_id BIGINT,
  tercero_id BIGINT NOT NULL, FOREIGN KEY (empresa_id, tercero_id) REFERENCES gv_cliente(empresa_id, id) ,
  moneda_id BIGINT NOT NULL REFERENCES gs_moneda(id), cambio_base factor NOT NULL DEFAULT 1,
  fecha_operacion TIMESTAMPTZ NOT NULL DEFAULT now(), confirmado_en TIMESTAMPTZ,
  estado TEXT NOT NULL DEFAULT 'BORRADOR' CHECK(estado IN ('BORRADOR','CONFIRMADO','REVERTIDO')),
  efecto_stock_aplicado TEXT CHECK(efecto_stock_aplicado IN ('ENTRADA','SALIDA','NINGUNO')),
  motivo_excepcion TEXT, total importe NOT NULL DEFAULT 0 CHECK(total>=0), tercero_snapshot objeto_json NOT NULL DEFAULT '{}',
  
  origen TEXT NOT NULL DEFAULT 'VENTAS' CHECK(origen IN ('VENTAS','POS')), costo_local NUMERIC(28,10),
  
  FOREIGN KEY (empresa_id, sucursal_id) REFERENCES gs_sucursal(empresa_id, id) , FOREIGN KEY (empresa_id, tipo_id) REFERENCES gs_tipo_comprobante(empresa_id, id) , FOREIGN KEY (empresa_id, numero_id) REFERENCES gs_talonario_numero(empresa_id, id) , FOREIGN KEY (empresa_id, terminal_id) REFERENCES gp_terminal(empresa_id, id) ,
  FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_comprobante_detalle (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, renglon INTEGER NOT NULL CHECK(renglon>0),
  articulo_id BIGINT NOT NULL, presentacion_id BIGINT NOT NULL, descripcion TEXT NOT NULL,
  cantidad cantidad NOT NULL CHECK(cantidad>0), factor_base factor NOT NULL,
  cantidad_base cantidad GENERATED ALWAYS AS(cantidad*factor_base) STORED,
  precio importe NOT NULL CHECK(precio>=0), descuento importe NOT NULL DEFAULT 0 CHECK(descuento>=0 AND descuento<=cantidad*precio),
  tasa_impuesto NUMERIC(9,6) NOT NULL DEFAULT 0 CHECK(tasa_impuesto BETWEEN 0 AND 100),
  base_imponible importe NOT NULL DEFAULT 0 CHECK(base_imponible>=0), impuesto importe NOT NULL DEFAULT 0 CHECK(impuesto>=0),
  importe_total importe GENERATED ALWAYS AS(cantidad*precio-descuento+impuesto) STORED,
  costo_unitario_base NUMERIC(28,10) CHECK(costo_unitario_base>=0), posicion_id BIGINT, destino_id BIGINT,
  FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gv_comprobante_cabecera(empresa_id, id) , FOREIGN KEY (empresa_id, articulo_id) REFERENCES gi_articulo(empresa_id, id) ,
  FOREIGN KEY(empresa_id,presentacion_id,articulo_id) REFERENCES gi_presentacion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,posicion_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,destino_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  UNIQUE(empresa_id,cabecera_id,renglon), UNIQUE(empresa_id,id,articulo_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gi_ajuste_cabecera (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  uid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE, sucursal_id BIGINT NOT NULL,
  usuario_id BIGINT NOT NULL, tipo_id BIGINT NOT NULL, numero_id BIGINT UNIQUE, terminal_id BIGINT,
  
  moneda_id BIGINT NOT NULL REFERENCES gs_moneda(id), cambio_base factor NOT NULL DEFAULT 1,
  fecha_operacion TIMESTAMPTZ NOT NULL DEFAULT now(), confirmado_en TIMESTAMPTZ,
  estado TEXT NOT NULL DEFAULT 'BORRADOR' CHECK(estado IN ('BORRADOR','CONFIRMADO','REVERTIDO')),
  efecto_stock_aplicado TEXT CHECK(efecto_stock_aplicado IN ('ENTRADA','SALIDA','NINGUNO')),
  motivo_excepcion TEXT, total importe NOT NULL DEFAULT 0 CHECK(total>=0), tercero_snapshot objeto_json NOT NULL DEFAULT '{}',
  
  
  
  FOREIGN KEY (empresa_id, sucursal_id) REFERENCES gs_sucursal(empresa_id, id) , FOREIGN KEY (empresa_id, tipo_id) REFERENCES gs_tipo_comprobante(empresa_id, id) , FOREIGN KEY (empresa_id, numero_id) REFERENCES gs_talonario_numero(empresa_id, id) , FOREIGN KEY (empresa_id, terminal_id) REFERENCES gp_terminal(empresa_id, id) ,
  FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gi_ajuste_detalle (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, renglon INTEGER NOT NULL CHECK(renglon>0),
  articulo_id BIGINT NOT NULL, presentacion_id BIGINT NOT NULL, descripcion TEXT NOT NULL,
  cantidad cantidad NOT NULL CHECK(cantidad>0), factor_base factor NOT NULL,
  cantidad_base cantidad GENERATED ALWAYS AS(cantidad*factor_base) STORED,
  precio importe NOT NULL CHECK(precio>=0), descuento importe NOT NULL DEFAULT 0 CHECK(descuento>=0 AND descuento<=cantidad*precio),
  tasa_impuesto NUMERIC(9,6) NOT NULL DEFAULT 0 CHECK(tasa_impuesto BETWEEN 0 AND 100),
  base_imponible importe NOT NULL DEFAULT 0 CHECK(base_imponible>=0), impuesto importe NOT NULL DEFAULT 0 CHECK(impuesto>=0),
  importe_total importe GENERATED ALWAYS AS(cantidad*precio-descuento+impuesto) STORED,
  costo_unitario_base NUMERIC(28,10) CHECK(costo_unitario_base>=0), posicion_id BIGINT, destino_id BIGINT,
  FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gi_ajuste_cabecera(empresa_id, id) , FOREIGN KEY (empresa_id, articulo_id) REFERENCES gi_articulo(empresa_id, id) ,
  FOREIGN KEY(empresa_id,presentacion_id,articulo_id) REFERENCES gi_presentacion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,posicion_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,destino_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  UNIQUE(empresa_id,cabecera_id,renglon), UNIQUE(empresa_id,id,articulo_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gi_transferencia_cabecera (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  uid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE, sucursal_id BIGINT NOT NULL,
  usuario_id BIGINT NOT NULL, tipo_id BIGINT NOT NULL, numero_id BIGINT UNIQUE, terminal_id BIGINT,
  
  moneda_id BIGINT NOT NULL REFERENCES gs_moneda(id), cambio_base factor NOT NULL DEFAULT 1,
  fecha_operacion TIMESTAMPTZ NOT NULL DEFAULT now(), confirmado_en TIMESTAMPTZ,
  estado TEXT NOT NULL DEFAULT 'BORRADOR' CHECK(estado IN ('BORRADOR','CONFIRMADO','REVERTIDO')),
  efecto_stock_aplicado TEXT CHECK(efecto_stock_aplicado IN ('ENTRADA','SALIDA','NINGUNO')),
  motivo_excepcion TEXT, total importe NOT NULL DEFAULT 0 CHECK(total>=0), tercero_snapshot objeto_json NOT NULL DEFAULT '{}',
  
  
  
  FOREIGN KEY (empresa_id, sucursal_id) REFERENCES gs_sucursal(empresa_id, id) , FOREIGN KEY (empresa_id, tipo_id) REFERENCES gs_tipo_comprobante(empresa_id, id) , FOREIGN KEY (empresa_id, numero_id) REFERENCES gs_talonario_numero(empresa_id, id) , FOREIGN KEY (empresa_id, terminal_id) REFERENCES gp_terminal(empresa_id, id) ,
  FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gi_transferencia_detalle (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, renglon INTEGER NOT NULL CHECK(renglon>0),
  articulo_id BIGINT NOT NULL, presentacion_id BIGINT NOT NULL, descripcion TEXT NOT NULL,
  cantidad cantidad NOT NULL CHECK(cantidad>0), factor_base factor NOT NULL,
  cantidad_base cantidad GENERATED ALWAYS AS(cantidad*factor_base) STORED,
  precio importe NOT NULL CHECK(precio>=0), descuento importe NOT NULL DEFAULT 0 CHECK(descuento>=0 AND descuento<=cantidad*precio),
  tasa_impuesto NUMERIC(9,6) NOT NULL DEFAULT 0 CHECK(tasa_impuesto BETWEEN 0 AND 100),
  base_imponible importe NOT NULL DEFAULT 0 CHECK(base_imponible>=0), impuesto importe NOT NULL DEFAULT 0 CHECK(impuesto>=0),
  importe_total importe GENERATED ALWAYS AS(cantidad*precio-descuento+impuesto) STORED,
  costo_unitario_base NUMERIC(28,10) CHECK(costo_unitario_base>=0), posicion_id BIGINT, destino_id BIGINT,
  FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gi_transferencia_cabecera(empresa_id, id) , FOREIGN KEY (empresa_id, articulo_id) REFERENCES gi_articulo(empresa_id, id) ,
  FOREIGN KEY(empresa_id,presentacion_id,articulo_id) REFERENCES gi_presentacion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,posicion_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,destino_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  UNIQUE(empresa_id,cabecera_id,renglon), UNIQUE(empresa_id,id,articulo_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gl_despacho_cabecera (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  uid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE, sucursal_id BIGINT NOT NULL,
  usuario_id BIGINT NOT NULL, tipo_id BIGINT NOT NULL, numero_id BIGINT UNIQUE, terminal_id BIGINT,
  
  moneda_id BIGINT NOT NULL REFERENCES gs_moneda(id), cambio_base factor NOT NULL DEFAULT 1,
  fecha_operacion TIMESTAMPTZ NOT NULL DEFAULT now(), confirmado_en TIMESTAMPTZ,
  estado TEXT NOT NULL DEFAULT 'BORRADOR' CHECK(estado IN ('BORRADOR','CONFIRMADO','REVERTIDO')),
  efecto_stock_aplicado TEXT CHECK(efecto_stock_aplicado IN ('ENTRADA','SALIDA','NINGUNO')),
  motivo_excepcion TEXT, total importe NOT NULL DEFAULT 0 CHECK(total>=0), tercero_snapshot objeto_json NOT NULL DEFAULT '{}',
  
  
  
  FOREIGN KEY (empresa_id, sucursal_id) REFERENCES gs_sucursal(empresa_id, id) , FOREIGN KEY (empresa_id, tipo_id) REFERENCES gs_tipo_comprobante(empresa_id, id) , FOREIGN KEY (empresa_id, numero_id) REFERENCES gs_talonario_numero(empresa_id, id) , FOREIGN KEY (empresa_id, terminal_id) REFERENCES gp_terminal(empresa_id, id) ,
  FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gl_despacho_detalle (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, renglon INTEGER NOT NULL CHECK(renglon>0),
  articulo_id BIGINT NOT NULL, presentacion_id BIGINT NOT NULL, descripcion TEXT NOT NULL,
  cantidad cantidad NOT NULL CHECK(cantidad>0), factor_base factor NOT NULL,
  cantidad_base cantidad GENERATED ALWAYS AS(cantidad*factor_base) STORED,
  precio importe NOT NULL CHECK(precio>=0), descuento importe NOT NULL DEFAULT 0 CHECK(descuento>=0 AND descuento<=cantidad*precio),
  tasa_impuesto NUMERIC(9,6) NOT NULL DEFAULT 0 CHECK(tasa_impuesto BETWEEN 0 AND 100),
  base_imponible importe NOT NULL DEFAULT 0 CHECK(base_imponible>=0), impuesto importe NOT NULL DEFAULT 0 CHECK(impuesto>=0),
  importe_total importe GENERATED ALWAYS AS(cantidad*precio-descuento+impuesto) STORED,
  costo_unitario_base NUMERIC(28,10) CHECK(costo_unitario_base>=0), posicion_id BIGINT, destino_id BIGINT,
  FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gl_despacho_cabecera(empresa_id, id) , FOREIGN KEY (empresa_id, articulo_id) REFERENCES gi_articulo(empresa_id, id) ,
  FOREIGN KEY(empresa_id,presentacion_id,articulo_id) REFERENCES gi_presentacion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,posicion_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,destino_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  UNIQUE(empresa_id,cabecera_id,renglon), UNIQUE(empresa_id,id,articulo_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_nota_credito_cabecera (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  uid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE, sucursal_id BIGINT NOT NULL,
  usuario_id BIGINT NOT NULL, tipo_id BIGINT NOT NULL, numero_id BIGINT UNIQUE, terminal_id BIGINT,
  tercero_id BIGINT NOT NULL, FOREIGN KEY (empresa_id, tercero_id) REFERENCES gv_cliente(empresa_id, id) ,
  moneda_id BIGINT NOT NULL REFERENCES gs_moneda(id), cambio_base factor NOT NULL DEFAULT 1,
  fecha_operacion TIMESTAMPTZ NOT NULL DEFAULT now(), confirmado_en TIMESTAMPTZ,
  estado TEXT NOT NULL DEFAULT 'BORRADOR' CHECK(estado IN ('BORRADOR','CONFIRMADO','REVERTIDO')),
  efecto_stock_aplicado TEXT CHECK(efecto_stock_aplicado IN ('ENTRADA','SALIDA','NINGUNO')),
  motivo_excepcion TEXT, total importe NOT NULL DEFAULT 0 CHECK(total>=0), tercero_snapshot objeto_json NOT NULL DEFAULT '{}',
  
  
  comprobante_origen_id BIGINT NOT NULL, FOREIGN KEY (empresa_id, comprobante_origen_id) REFERENCES gv_comprobante_cabecera(empresa_id, id) ,
  FOREIGN KEY (empresa_id, sucursal_id) REFERENCES gs_sucursal(empresa_id, id) , FOREIGN KEY (empresa_id, tipo_id) REFERENCES gs_tipo_comprobante(empresa_id, id) , FOREIGN KEY (empresa_id, numero_id) REFERENCES gs_talonario_numero(empresa_id, id) , FOREIGN KEY (empresa_id, terminal_id) REFERENCES gp_terminal(empresa_id, id) ,
  FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_nota_credito_detalle (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, renglon INTEGER NOT NULL CHECK(renglon>0),
  articulo_id BIGINT NOT NULL, presentacion_id BIGINT NOT NULL, descripcion TEXT NOT NULL,
  cantidad cantidad NOT NULL CHECK(cantidad>0), factor_base factor NOT NULL,
  cantidad_base cantidad GENERATED ALWAYS AS(cantidad*factor_base) STORED,
  precio importe NOT NULL CHECK(precio>=0), descuento importe NOT NULL DEFAULT 0 CHECK(descuento>=0 AND descuento<=cantidad*precio),
  tasa_impuesto NUMERIC(9,6) NOT NULL DEFAULT 0 CHECK(tasa_impuesto BETWEEN 0 AND 100),
  base_imponible importe NOT NULL DEFAULT 0 CHECK(base_imponible>=0), impuesto importe NOT NULL DEFAULT 0 CHECK(impuesto>=0),
  importe_total importe GENERATED ALWAYS AS(cantidad*precio-descuento+impuesto) STORED,
  costo_unitario_base NUMERIC(28,10) CHECK(costo_unitario_base>=0), posicion_id BIGINT, destino_id BIGINT,
  FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gv_nota_credito_cabecera(empresa_id, id) , FOREIGN KEY (empresa_id, articulo_id) REFERENCES gi_articulo(empresa_id, id) ,
  FOREIGN KEY(empresa_id,presentacion_id,articulo_id) REFERENCES gi_presentacion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,posicion_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,destino_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  UNIQUE(empresa_id,cabecera_id,renglon), UNIQUE(empresa_id,id,articulo_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gc_nota_credito_cabecera (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  uid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE, sucursal_id BIGINT NOT NULL,
  usuario_id BIGINT NOT NULL, tipo_id BIGINT NOT NULL, numero_id BIGINT UNIQUE, terminal_id BIGINT,
  tercero_id BIGINT NOT NULL, FOREIGN KEY (empresa_id, tercero_id) REFERENCES gc_proveedor(empresa_id, id) ,
  moneda_id BIGINT NOT NULL REFERENCES gs_moneda(id), cambio_base factor NOT NULL DEFAULT 1,
  fecha_operacion TIMESTAMPTZ NOT NULL DEFAULT now(), confirmado_en TIMESTAMPTZ,
  estado TEXT NOT NULL DEFAULT 'BORRADOR' CHECK(estado IN ('BORRADOR','CONFIRMADO','REVERTIDO')),
  efecto_stock_aplicado TEXT CHECK(efecto_stock_aplicado IN ('ENTRADA','SALIDA','NINGUNO')),
  motivo_excepcion TEXT, total importe NOT NULL DEFAULT 0 CHECK(total>=0), tercero_snapshot objeto_json NOT NULL DEFAULT '{}',
  numero_proveedor TEXT NOT NULL, timbrado_proveedor TEXT NOT NULL DEFAULT '', UNIQUE(empresa_id,tercero_id,tipo_id,timbrado_proveedor,numero_proveedor),
  
  comprobante_origen_id BIGINT NOT NULL, FOREIGN KEY (empresa_id, comprobante_origen_id) REFERENCES gc_comprobante_cabecera(empresa_id, id) ,
  FOREIGN KEY (empresa_id, sucursal_id) REFERENCES gs_sucursal(empresa_id, id) , FOREIGN KEY (empresa_id, tipo_id) REFERENCES gs_tipo_comprobante(empresa_id, id) , FOREIGN KEY (empresa_id, numero_id) REFERENCES gs_talonario_numero(empresa_id, id) , FOREIGN KEY (empresa_id, terminal_id) REFERENCES gp_terminal(empresa_id, id) ,
  FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gc_nota_credito_detalle (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, renglon INTEGER NOT NULL CHECK(renglon>0),
  articulo_id BIGINT NOT NULL, presentacion_id BIGINT NOT NULL, descripcion TEXT NOT NULL,
  cantidad cantidad NOT NULL CHECK(cantidad>0), factor_base factor NOT NULL,
  cantidad_base cantidad GENERATED ALWAYS AS(cantidad*factor_base) STORED,
  precio importe NOT NULL CHECK(precio>=0), descuento importe NOT NULL DEFAULT 0 CHECK(descuento>=0 AND descuento<=cantidad*precio),
  tasa_impuesto NUMERIC(9,6) NOT NULL DEFAULT 0 CHECK(tasa_impuesto BETWEEN 0 AND 100),
  base_imponible importe NOT NULL DEFAULT 0 CHECK(base_imponible>=0), impuesto importe NOT NULL DEFAULT 0 CHECK(impuesto>=0),
  importe_total importe GENERATED ALWAYS AS(cantidad*precio-descuento+impuesto) STORED,
  costo_unitario_base NUMERIC(28,10) CHECK(costo_unitario_base>=0), posicion_id BIGINT, destino_id BIGINT,
  FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gc_nota_credito_cabecera(empresa_id, id) , FOREIGN KEY (empresa_id, articulo_id) REFERENCES gi_articulo(empresa_id, id) ,
  FOREIGN KEY(empresa_id,presentacion_id,articulo_id) REFERENCES gi_presentacion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,posicion_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  FOREIGN KEY(empresa_id,destino_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
  UNIQUE(empresa_id,cabecera_id,renglon), UNIQUE(empresa_id,id,articulo_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gc_recepcion_orden (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  recepcion_detalle_id BIGINT NOT NULL, orden_detalle_id BIGINT NOT NULL, cantidad_base cantidad NOT NULL CHECK(cantidad_base>0),
 FOREIGN KEY (empresa_id, recepcion_detalle_id) REFERENCES gc_recepcion_detalle(empresa_id, id) , FOREIGN KEY (empresa_id, orden_detalle_id) REFERENCES gc_orden_compra_detalle(empresa_id, id) , UNIQUE(empresa_id,recepcion_detalle_id,orden_detalle_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gc_factura_recepcion_detalle (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  factura_detalle_id BIGINT NOT NULL, recepcion_detalle_id BIGINT NOT NULL,
 cantidad_base cantidad NOT NULL CHECK(cantidad_base>0), costo_final_base NUMERIC(28,10) NOT NULL CHECK(costo_final_base>=0),
 diferencia_base NUMERIC(28,10), aplicado BOOLEAN NOT NULL DEFAULT FALSE,
 FOREIGN KEY (empresa_id, factura_detalle_id) REFERENCES gc_comprobante_detalle(empresa_id, id) , FOREIGN KEY (empresa_id, recepcion_detalle_id) REFERENCES gc_recepcion_detalle(empresa_id, id) , UNIQUE(empresa_id,factura_detalle_id,recepcion_detalle_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_pedido_factura (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  pedido_detalle_id BIGINT NOT NULL, factura_detalle_id BIGINT NOT NULL, cantidad_base cantidad NOT NULL CHECK(cantidad_base>0),
 FOREIGN KEY (empresa_id, pedido_detalle_id) REFERENCES gv_pedido_detalle(empresa_id, id) , FOREIGN KEY (empresa_id, factura_detalle_id) REFERENCES gv_comprobante_detalle(empresa_id, id) , UNIQUE(empresa_id,pedido_detalle_id,factura_detalle_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gl_asignacion_venta (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  despacho_detalle_id BIGINT NOT NULL, venta_detalle_id BIGINT NOT NULL, cantidad_base cantidad NOT NULL CHECK(cantidad_base>0),
 FOREIGN KEY (empresa_id, despacho_detalle_id) REFERENCES gl_despacho_detalle(empresa_id, id) , FOREIGN KEY (empresa_id, venta_detalle_id) REFERENCES gv_comprobante_detalle(empresa_id, id) , UNIQUE(empresa_id,despacho_detalle_id,venta_detalle_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gl_entrega_evento (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  despacho_id BIGINT NOT NULL, fecha TIMESTAMPTZ NOT NULL DEFAULT now(),
 estado TEXT NOT NULL CHECK(estado IN ('PREPARADO','DESPACHADO','ENTREGADO','RECHAZADO')), receptor TEXT, evidencia TEXT,
 FOREIGN KEY (empresa_id, despacho_id) REFERENCES gl_despacho_cabecera(empresa_id, id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gi_movimiento (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  uid UUID NOT NULL UNIQUE, articulo_id BIGINT NOT NULL, posicion_id BIGINT,
 fecha_operacion TIMESTAMPTZ NOT NULL, registrado_en TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 cantidad cantidad NOT NULL, valor_inventario NUMERIC(28,10) NOT NULL, diferencia_consumida NUMERIC(28,10) NOT NULL DEFAULT 0,
 costo_unitario NUMERIC(28,10), cantidad_antes cantidad NOT NULL, cantidad_despues cantidad NOT NULL,
 valor_antes NUMERIC(28,10) NOT NULL, valor_despues NUMERIC(28,10) NOT NULL, valoracion_pendiente BOOLEAN NOT NULL DEFAULT FALSE,
 recepcion_detalle_id BIGINT, venta_detalle_id BIGINT, ajuste_detalle_id BIGINT, transferencia_detalle_id BIGINT,
 factura_recepcion_id BIGINT, reversion_de BIGINT, motivo TEXT NOT NULL,
 FOREIGN KEY (empresa_id, articulo_id) REFERENCES gi_articulo(empresa_id, id) , FOREIGN KEY(empresa_id,posicion_id,articulo_id) REFERENCES gi_posicion(empresa_id,id,articulo_id),
 FOREIGN KEY (empresa_id, recepcion_detalle_id) REFERENCES gc_recepcion_detalle(empresa_id, id) , FOREIGN KEY (empresa_id, venta_detalle_id) REFERENCES gv_comprobante_detalle(empresa_id, id) ,
 FOREIGN KEY (empresa_id, ajuste_detalle_id) REFERENCES gi_ajuste_detalle(empresa_id, id) , FOREIGN KEY (empresa_id, transferencia_detalle_id) REFERENCES gi_transferencia_detalle(empresa_id, id) ,
 FOREIGN KEY (empresa_id, factura_recepcion_id) REFERENCES gc_factura_recepcion_detalle(empresa_id, id) , FOREIGN KEY (empresa_id, reversion_de) REFERENCES gi_movimiento(empresa_id, id) ,
 CHECK(num_nonnulls(recepcion_detalle_id,venta_detalle_id,ajuste_detalle_id,transferencia_detalle_id,factura_recepcion_id,reversion_de)=1),
 CHECK(posicion_id IS NOT NULL OR cantidad=0), UNIQUE(empresa_id,reversion_de),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gi_incidencia (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  movimiento_id BIGINT NOT NULL, tipo TEXT NOT NULL CHECK(tipo IN ('NEGATIVO','SIN_COSTO','DIFERENCIA')),
 creada_en TIMESTAMPTZ NOT NULL DEFAULT now(), resuelta_en TIMESTAMPTZ, resolucion TEXT,
 FOREIGN KEY (empresa_id, movimiento_id) REFERENCES gi_movimiento(empresa_id, id) , CHECK(resuelta_en IS NULL OR resolucion IS NOT NULL),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gi_reserva (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  posicion_id BIGINT NOT NULL, pedido_detalle_id BIGINT, terminal_id BIGINT, cantidad cantidad NOT NULL CHECK(cantidad>0),
 consumida cantidad NOT NULL DEFAULT 0 CHECK(consumida>=0 AND consumida<=cantidad), activa BOOLEAN NOT NULL DEFAULT TRUE,
 FOREIGN KEY (empresa_id, posicion_id) REFERENCES gi_posicion(empresa_id, id) , FOREIGN KEY (empresa_id, pedido_detalle_id) REFERENCES gv_pedido_detalle(empresa_id, id) , FOREIGN KEY (empresa_id, terminal_id) REFERENCES gp_terminal(empresa_id, id) ,
 CHECK(num_nonnulls(pedido_detalle_id,terminal_id)=1),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_devolucion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  uid UUID NOT NULL UNIQUE, detalle_origen_id BIGINT NOT NULL,
 cantidad_base cantidad NOT NULL CHECK(cantidad_base>0), movimiento_id BIGINT, fecha TIMESTAMPTZ NOT NULL DEFAULT now(), motivo TEXT NOT NULL,
 FOREIGN KEY (empresa_id, detalle_origen_id) REFERENCES gv_comprobante_detalle(empresa_id, id) , FOREIGN KEY (empresa_id, movimiento_id) REFERENCES gi_movimiento(empresa_id, id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gc_devolucion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  uid UUID NOT NULL UNIQUE, detalle_origen_id BIGINT NOT NULL,
 cantidad_base cantidad NOT NULL CHECK(cantidad_base>0), movimiento_id BIGINT, fecha TIMESTAMPTZ NOT NULL DEFAULT now(), motivo TEXT NOT NULL,
 FOREIGN KEY (empresa_id, detalle_origen_id) REFERENCES gc_recepcion_detalle(empresa_id, id) , FOREIGN KEY (empresa_id, movimiento_id) REFERENCES gi_movimiento(empresa_id, id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gc_orden_compra_anulacion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, usuario_id BIGINT NOT NULL, motivo TEXT NOT NULL,
 fecha TIMESTAMPTZ NOT NULL DEFAULT now(), FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gc_orden_compra_cabecera(empresa_id, id) ,
 FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id), UNIQUE(empresa_id,cabecera_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gc_recepcion_anulacion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, usuario_id BIGINT NOT NULL, motivo TEXT NOT NULL,
 fecha TIMESTAMPTZ NOT NULL DEFAULT now(), FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gc_recepcion_cabecera(empresa_id, id) ,
 FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id), UNIQUE(empresa_id,cabecera_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gc_comprobante_anulacion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, usuario_id BIGINT NOT NULL, motivo TEXT NOT NULL,
 fecha TIMESTAMPTZ NOT NULL DEFAULT now(), FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gc_comprobante_cabecera(empresa_id, id) ,
 FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id), UNIQUE(empresa_id,cabecera_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_cotizacion_anulacion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, usuario_id BIGINT NOT NULL, motivo TEXT NOT NULL,
 fecha TIMESTAMPTZ NOT NULL DEFAULT now(), FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gv_cotizacion_cabecera(empresa_id, id) ,
 FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id), UNIQUE(empresa_id,cabecera_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_pedido_anulacion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, usuario_id BIGINT NOT NULL, motivo TEXT NOT NULL,
 fecha TIMESTAMPTZ NOT NULL DEFAULT now(), FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gv_pedido_cabecera(empresa_id, id) ,
 FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id), UNIQUE(empresa_id,cabecera_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_comprobante_anulacion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, usuario_id BIGINT NOT NULL, motivo TEXT NOT NULL,
 fecha TIMESTAMPTZ NOT NULL DEFAULT now(), FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gv_comprobante_cabecera(empresa_id, id) ,
 FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id), UNIQUE(empresa_id,cabecera_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gi_ajuste_anulacion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, usuario_id BIGINT NOT NULL, motivo TEXT NOT NULL,
 fecha TIMESTAMPTZ NOT NULL DEFAULT now(), FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gi_ajuste_cabecera(empresa_id, id) ,
 FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id), UNIQUE(empresa_id,cabecera_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gi_transferencia_anulacion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, usuario_id BIGINT NOT NULL, motivo TEXT NOT NULL,
 fecha TIMESTAMPTZ NOT NULL DEFAULT now(), FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gi_transferencia_cabecera(empresa_id, id) ,
 FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id), UNIQUE(empresa_id,cabecera_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gl_despacho_anulacion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, usuario_id BIGINT NOT NULL, motivo TEXT NOT NULL,
 fecha TIMESTAMPTZ NOT NULL DEFAULT now(), FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gl_despacho_cabecera(empresa_id, id) ,
 FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id), UNIQUE(empresa_id,cabecera_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_nota_credito_anulacion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, usuario_id BIGINT NOT NULL, motivo TEXT NOT NULL,
 fecha TIMESTAMPTZ NOT NULL DEFAULT now(), FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gv_nota_credito_cabecera(empresa_id, id) ,
 FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id), UNIQUE(empresa_id,cabecera_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gc_nota_credito_anulacion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, usuario_id BIGINT NOT NULL, motivo TEXT NOT NULL,
 fecha TIMESTAMPTZ NOT NULL DEFAULT now(), FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gc_nota_credito_cabecera(empresa_id, id) ,
 FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id), UNIQUE(empresa_id,cabecera_id),
  UNIQUE (empresa_id, id)
);

-- =====================================================================
-- 7. Caja, cuentas por cobrar/pagar y aplicaciones
-- =====================================================================

CREATE TABLE gf_caja (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  sucursal_id BIGINT NOT NULL, codigo TEXT NOT NULL, moneda_id BIGINT NOT NULL REFERENCES gs_moneda(id),
 FOREIGN KEY (empresa_id, sucursal_id) REFERENCES gs_sucursal(empresa_id, id) , UNIQUE(empresa_id,codigo),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gf_caja_apertura (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  caja_id BIGINT NOT NULL, usuario_id BIGINT NOT NULL, terminal_id BIGINT,
 abierta_en TIMESTAMPTZ NOT NULL DEFAULT now(), cerrada_en TIMESTAMPTZ, fondo importe NOT NULL DEFAULT 0 CHECK(fondo>=0),
 FOREIGN KEY (empresa_id, caja_id) REFERENCES gf_caja(empresa_id, id) , FOREIGN KEY (empresa_id, terminal_id) REFERENCES gp_terminal(empresa_id, id) , FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id), CHECK(cerrada_en IS NULL OR cerrada_en>=abierta_en),
  UNIQUE (empresa_id, id)
);

CREATE UNIQUE INDEX caja_abierta ON gf_caja_apertura(empresa_id,caja_id) WHERE cerrada_en IS NULL;

CREATE TABLE gf_medio_pago (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  codigo TEXT NOT NULL, nombre TEXT NOT NULL, efectivo BOOLEAN NOT NULL DEFAULT FALSE, UNIQUE(empresa_id,codigo),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_cuota (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  comprobante_id BIGINT NOT NULL, numero INTEGER NOT NULL CHECK(numero>0), vencimiento DATE NOT NULL,
 importe importe NOT NULL CHECK(importe>0), aplicado importe NOT NULL DEFAULT 0 CHECK(aplicado>=0 AND aplicado<=importe),
 creditado importe NOT NULL DEFAULT 0 CHECK(creditado>=0 AND creditado+aplicado<=importe),
 saldo importe GENERATED ALWAYS AS(importe-aplicado-creditado) STORED, FOREIGN KEY (empresa_id, comprobante_id) REFERENCES gv_comprobante_cabecera(empresa_id, id) , UNIQUE(empresa_id,comprobante_id,numero),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gc_cuota (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  comprobante_id BIGINT NOT NULL, numero INTEGER NOT NULL CHECK(numero>0), vencimiento DATE NOT NULL,
 importe importe NOT NULL CHECK(importe>0), aplicado importe NOT NULL DEFAULT 0 CHECK(aplicado>=0 AND aplicado<=importe),
 creditado importe NOT NULL DEFAULT 0 CHECK(creditado>=0 AND creditado+aplicado<=importe),
 saldo importe GENERATED ALWAYS AS(importe-aplicado-creditado) STORED, FOREIGN KEY (empresa_id, comprobante_id) REFERENCES gc_comprobante_cabecera(empresa_id, id) , UNIQUE(empresa_id,comprobante_id,numero),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_credito_aplicacion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  nota_id BIGINT NOT NULL, cuota_id BIGINT NOT NULL, importe importe NOT NULL CHECK(importe>0),
 FOREIGN KEY (empresa_id, nota_id) REFERENCES gv_nota_credito_cabecera(empresa_id, id) , FOREIGN KEY (empresa_id, cuota_id) REFERENCES gv_cuota(empresa_id, id) , UNIQUE(empresa_id,nota_id,cuota_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_credito_disponible (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  nota_id BIGINT NOT NULL, importe importe NOT NULL CHECK(importe>0), utilizado importe NOT NULL DEFAULT 0 CHECK(utilizado>=0 AND utilizado<=importe),
 saldo importe GENERATED ALWAYS AS(importe-utilizado) STORED, FOREIGN KEY (empresa_id, nota_id) REFERENCES gv_nota_credito_cabecera(empresa_id, id) , UNIQUE(empresa_id,nota_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gv_credito_uso (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  uid UUID NOT NULL UNIQUE, credito_id BIGINT NOT NULL, cuota_id BIGINT NOT NULL, importe importe NOT NULL CHECK(importe>0),
 FOREIGN KEY (empresa_id, credito_id) REFERENCES gv_credito_disponible(empresa_id, id) , FOREIGN KEY (empresa_id, cuota_id) REFERENCES gv_cuota(empresa_id, id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gc_credito_aplicacion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  nota_id BIGINT NOT NULL, cuota_id BIGINT NOT NULL, importe importe NOT NULL CHECK(importe>0),
 FOREIGN KEY (empresa_id, nota_id) REFERENCES gc_nota_credito_cabecera(empresa_id, id) , FOREIGN KEY (empresa_id, cuota_id) REFERENCES gc_cuota(empresa_id, id) , UNIQUE(empresa_id,nota_id,cuota_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gc_credito_disponible (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  nota_id BIGINT NOT NULL, importe importe NOT NULL CHECK(importe>0), utilizado importe NOT NULL DEFAULT 0 CHECK(utilizado>=0 AND utilizado<=importe),
 saldo importe GENERATED ALWAYS AS(importe-utilizado) STORED, FOREIGN KEY (empresa_id, nota_id) REFERENCES gc_nota_credito_cabecera(empresa_id, id) , UNIQUE(empresa_id,nota_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gc_credito_uso (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  uid UUID NOT NULL UNIQUE, credito_id BIGINT NOT NULL, cuota_id BIGINT NOT NULL, importe importe NOT NULL CHECK(importe>0),
 FOREIGN KEY (empresa_id, credito_id) REFERENCES gc_credito_disponible(empresa_id, id) , FOREIGN KEY (empresa_id, cuota_id) REFERENCES gc_cuota(empresa_id, id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gf_recibo_cabecera (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  uid UUID NOT NULL UNIQUE, tercero_id BIGINT NOT NULL, apertura_id BIGINT, numero_id BIGINT UNIQUE,
 moneda_id BIGINT NOT NULL REFERENCES gs_moneda(id), importe importe NOT NULL CHECK(importe>0),
 estado TEXT NOT NULL DEFAULT 'BORRADOR' CHECK(estado IN ('BORRADOR','CONFIRMADO','REVERTIDO')), fecha TIMESTAMPTZ NOT NULL DEFAULT now(),
 FOREIGN KEY (empresa_id, tercero_id) REFERENCES gv_cliente(empresa_id, id) , FOREIGN KEY (empresa_id, apertura_id) REFERENCES gf_caja_apertura(empresa_id, id) , FOREIGN KEY (empresa_id, numero_id) REFERENCES gs_talonario_numero(empresa_id, id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gf_recibo_forma (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, medio_id BIGINT NOT NULL, importe importe NOT NULL CHECK(importe>0), referencia TEXT,
 FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gf_recibo_cabecera(empresa_id, id) , FOREIGN KEY (empresa_id, medio_id) REFERENCES gf_medio_pago(empresa_id, id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gf_recibo_aplicacion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, cuota_id BIGINT NOT NULL, importe importe NOT NULL CHECK(importe>0),
 FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gf_recibo_cabecera(empresa_id, id) , FOREIGN KEY (empresa_id, cuota_id) REFERENCES gv_cuota(empresa_id, id) , UNIQUE(empresa_id,cabecera_id,cuota_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gf_pago_cabecera (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  uid UUID NOT NULL UNIQUE, tercero_id BIGINT NOT NULL, apertura_id BIGINT, numero_id BIGINT UNIQUE,
 moneda_id BIGINT NOT NULL REFERENCES gs_moneda(id), importe importe NOT NULL CHECK(importe>0),
 estado TEXT NOT NULL DEFAULT 'BORRADOR' CHECK(estado IN ('BORRADOR','CONFIRMADO','REVERTIDO')), fecha TIMESTAMPTZ NOT NULL DEFAULT now(),
 FOREIGN KEY (empresa_id, tercero_id) REFERENCES gc_proveedor(empresa_id, id) , FOREIGN KEY (empresa_id, apertura_id) REFERENCES gf_caja_apertura(empresa_id, id) , FOREIGN KEY (empresa_id, numero_id) REFERENCES gs_talonario_numero(empresa_id, id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gf_pago_forma (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, medio_id BIGINT NOT NULL, importe importe NOT NULL CHECK(importe>0), referencia TEXT,
 FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gf_pago_cabecera(empresa_id, id) , FOREIGN KEY (empresa_id, medio_id) REFERENCES gf_medio_pago(empresa_id, id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gf_pago_aplicacion (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cabecera_id BIGINT NOT NULL, cuota_id BIGINT NOT NULL, importe importe NOT NULL CHECK(importe>0),
 FOREIGN KEY (empresa_id, cabecera_id) REFERENCES gf_pago_cabecera(empresa_id, id) , FOREIGN KEY (empresa_id, cuota_id) REFERENCES gc_cuota(empresa_id, id) , UNIQUE(empresa_id,cabecera_id,cuota_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gf_caja_arqueo (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  apertura_id BIGINT NOT NULL, medio_id BIGINT NOT NULL, contado importe NOT NULL CHECK(contado>=0), esperado importe NOT NULL,
 diferencia importe GENERATED ALWAYS AS(contado-esperado) STORED, FOREIGN KEY (empresa_id, apertura_id) REFERENCES gf_caja_apertura(empresa_id, id) , FOREIGN KEY (empresa_id, medio_id) REFERENCES gf_medio_pago(empresa_id, id) , UNIQUE(empresa_id,apertura_id,medio_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gf_cuenta_bancaria (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  codigo TEXT NOT NULL, banco TEXT NOT NULL, numero_cuenta TEXT NOT NULL, moneda_id BIGINT NOT NULL REFERENCES gs_moneda(id), UNIQUE(empresa_id,codigo),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gf_movimiento_bancario (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  cuenta_id BIGINT NOT NULL, uid UUID NOT NULL UNIQUE, fecha DATE NOT NULL, importe importe NOT NULL CHECK(importe<>0),
 referencia TEXT NOT NULL, conciliado_en TIMESTAMPTZ, FOREIGN KEY (empresa_id, cuenta_id) REFERENCES gf_cuenta_bancaria(empresa_id, id),
  UNIQUE (empresa_id, id)
);

-- =====================================================================
-- 8. Offline, fiscal y auditoria
-- =====================================================================

CREATE TABLE gp_permiso_offline (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  terminal_id BIGINT NOT NULL, usuario_id BIGINT NOT NULL, emitido_en TIMESTAMPTZ NOT NULL DEFAULT now(),
 vence_en TIMESTAMPTZ NOT NULL, supervisor BOOLEAN NOT NULL DEFAULT FALSE, firma TEXT NOT NULL,
 FOREIGN KEY (empresa_id, terminal_id) REFERENCES gp_terminal(empresa_id, id) , FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id), CHECK(vence_en>emitido_en),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gp_extension_permiso (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  permiso_id BIGINT NOT NULL, supervisor_permiso_id BIGINT NOT NULL, hasta TIMESTAMPTZ NOT NULL,
 motivo TEXT NOT NULL, uid UUID NOT NULL UNIQUE, FOREIGN KEY (empresa_id, permiso_id) REFERENCES gp_permiso_offline(empresa_id, id) , FOREIGN KEY (empresa_id, supervisor_permiso_id) REFERENCES gp_permiso_offline(empresa_id, id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gp_evento_entrada (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  uid UUID NOT NULL UNIQUE, terminal_id BIGINT NOT NULL, secuencia BIGINT NOT NULL CHECK(secuencia>0),
 ocurrido_en TIMESTAMPTZ NOT NULL, recibido_en TIMESTAMPTZ NOT NULL DEFAULT now(), contenido objeto_json NOT NULL,
 venta_id BIGINT, estado TEXT NOT NULL DEFAULT 'PENDIENTE' CHECK(estado IN ('PENDIENTE','APLICADO','CONFLICTO')),
 error TEXT, FOREIGN KEY (empresa_id, terminal_id) REFERENCES gp_terminal(empresa_id, id) , FOREIGN KEY (empresa_id, venta_id) REFERENCES gv_comprobante_cabecera(empresa_id, id) , UNIQUE(empresa_id,terminal_id,secuencia),
  UNIQUE (empresa_id, id)
);

CREATE TABLE gs_evento_salida (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  uid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE, tema TEXT NOT NULL, contenido objeto_json NOT NULL,
 creado_en TIMESTAMPTZ NOT NULL DEFAULT now(), publicado_en TIMESTAMPTZ,
  UNIQUE (empresa_id, id)
);

CREATE TABLE gs_auditoria (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  usuario_id BIGINT, fecha TIMESTAMPTZ NOT NULL DEFAULT now(), tabla TEXT NOT NULL, registro_id BIGINT NOT NULL,
 accion TEXT NOT NULL, antes JSONB, despues JSONB, FOREIGN KEY(empresa_id,usuario_id) REFERENCES gs_usuario_empresa(empresa_id,usuario_id),
  UNIQUE (empresa_id, id)
);

CREATE TABLE fe_documento (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  numero_id BIGINT NOT NULL, cdc VARCHAR(44), estado TEXT NOT NULL DEFAULT 'PENDIENTE' CHECK(estado IN ('PENDIENTE','FIRMADO','ENVIADO','APROBADO','RECHAZADO','CANCELADO')),
 xml_firmado TEXT, qr TEXT, version_esquema TEXT, FOREIGN KEY (empresa_id, numero_id) REFERENCES gs_talonario_numero(empresa_id, id) , UNIQUE(empresa_id,numero_id), UNIQUE(cdc), CHECK(cdc IS NULL OR cdc ~ '^[0-9]{44}$'),
  UNIQUE (empresa_id, id)
);

CREATE TABLE fe_evento (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  empresa_id BIGINT NOT NULL REFERENCES gs_empresa(id),
  documento_id BIGINT NOT NULL, uid UUID NOT NULL UNIQUE, tipo TEXT NOT NULL, fecha TIMESTAMPTZ NOT NULL DEFAULT now(),
 solicitud JSONB, respuesta JSONB, FOREIGN KEY (empresa_id, documento_id) REFERENCES fe_documento(empresa_id, id),
  UNIQUE (empresa_id, id)
);

COMMENT ON TABLE gs_moneda IS 'gs_moneda: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_empresa IS 'gs_empresa: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_usuario IS 'gs_usuario: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_usuario_empresa IS 'gs_usuario_empresa: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_sucursal IS 'gs_sucursal: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_rol IS 'gs_rol: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_permiso IS 'gs_permiso: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_rol_permiso IS 'gs_rol_permiso: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_usuario_rol IS 'gs_usuario_rol: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_modulo IS 'gs_modulo: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_empresa_modulo IS 'gs_empresa_modulo: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_cotizacion_moneda IS 'gs_cotizacion_moneda: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_entidad IS 'gs_entidad: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_entidad_documento IS 'gs_entidad_documento: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_entidad_contacto IS 'gs_entidad_contacto: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_entidad_direccion IS 'gs_entidad_direccion: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gf_condicion_pago IS 'gf_condicion_pago: entidad del modulo gf; ver diccionario V4.';

COMMENT ON TABLE gf_condicion_pago_cuota IS 'gf_condicion_pago_cuota: entidad del modulo gf; ver diccionario V4.';

COMMENT ON TABLE gv_cliente IS 'gv_cliente: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gc_proveedor IS 'gc_proveedor: entidad del modulo gc; ver diccionario V4.';

COMMENT ON TABLE gi_unidad_medida IS 'gi_unidad_medida: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gi_categoria IS 'gi_categoria: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gi_producto IS 'gi_producto: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gi_articulo IS 'gi_articulo: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gi_presentacion IS 'gi_presentacion: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gi_codigo_barra IS 'gi_codigo_barra: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gp_regla_balanza IS 'gp_regla_balanza: entidad del modulo gp; ver diccionario V4.';

COMMENT ON TABLE gi_kit_componente IS 'gi_kit_componente: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gi_lote IS 'gi_lote: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gi_serie IS 'gi_serie: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gs_impuesto IS 'gs_impuesto: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_impuesto_tasa IS 'gs_impuesto_tasa: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gi_articulo_impuesto IS 'gi_articulo_impuesto: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gv_lista_precio IS 'gv_lista_precio: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gv_precio IS 'gv_precio: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gv_promocion IS 'gv_promocion: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gv_promocion_articulo IS 'gv_promocion_articulo: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gs_usuario_sucursal IS 'gs_usuario_sucursal: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gp_terminal IS 'gp_terminal: entidad del modulo gp; ver diccionario V4.';

COMMENT ON TABLE gs_modalidad_emision IS 'gs_modalidad_emision: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_habilitacion_fiscal IS 'gs_habilitacion_fiscal: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_clase_documento IS 'gs_clase_documento: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_tipo_comprobante IS 'gs_tipo_comprobante: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_talonario IS 'gs_talonario: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_talonario_usuario IS 'gs_talonario_usuario: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_talonario_alternativo IS 'gs_talonario_alternativo: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_talonario_rango_terminal IS 'gs_talonario_rango_terminal: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_talonario_numero IS 'gs_talonario_numero: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gi_deposito IS 'gi_deposito: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gi_ubicacion IS 'gi_ubicacion: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gi_posicion IS 'gi_posicion: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gi_existencia IS 'gi_existencia: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gi_valoracion IS 'gi_valoracion: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gc_orden_compra_cabecera IS 'gc_orden_compra_cabecera: entidad del modulo gc; ver diccionario V4.';

COMMENT ON TABLE gc_orden_compra_detalle IS 'gc_orden_compra_detalle: entidad del modulo gc; ver diccionario V4.';

COMMENT ON TABLE gc_recepcion_cabecera IS 'gc_recepcion_cabecera: entidad del modulo gc; ver diccionario V4.';

COMMENT ON TABLE gc_recepcion_detalle IS 'gc_recepcion_detalle: entidad del modulo gc; ver diccionario V4.';

COMMENT ON TABLE gc_comprobante_cabecera IS 'gc_comprobante_cabecera: entidad del modulo gc; ver diccionario V4.';

COMMENT ON TABLE gc_comprobante_detalle IS 'gc_comprobante_detalle: entidad del modulo gc; ver diccionario V4.';

COMMENT ON TABLE gv_cotizacion_cabecera IS 'gv_cotizacion_cabecera: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gv_cotizacion_detalle IS 'gv_cotizacion_detalle: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gv_pedido_cabecera IS 'gv_pedido_cabecera: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gv_pedido_detalle IS 'gv_pedido_detalle: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gv_comprobante_cabecera IS 'gv_comprobante_cabecera: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gv_comprobante_detalle IS 'gv_comprobante_detalle: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gi_ajuste_cabecera IS 'gi_ajuste_cabecera: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gi_ajuste_detalle IS 'gi_ajuste_detalle: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gi_transferencia_cabecera IS 'gi_transferencia_cabecera: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gi_transferencia_detalle IS 'gi_transferencia_detalle: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gl_despacho_cabecera IS 'gl_despacho_cabecera: entidad del modulo gl; ver diccionario V4.';

COMMENT ON TABLE gl_despacho_detalle IS 'gl_despacho_detalle: entidad del modulo gl; ver diccionario V4.';

COMMENT ON TABLE gv_nota_credito_cabecera IS 'gv_nota_credito_cabecera: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gv_nota_credito_detalle IS 'gv_nota_credito_detalle: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gc_nota_credito_cabecera IS 'gc_nota_credito_cabecera: entidad del modulo gc; ver diccionario V4.';

COMMENT ON TABLE gc_nota_credito_detalle IS 'gc_nota_credito_detalle: entidad del modulo gc; ver diccionario V4.';

COMMENT ON TABLE gc_recepcion_orden IS 'gc_recepcion_orden: entidad del modulo gc; ver diccionario V4.';

COMMENT ON TABLE gc_factura_recepcion_detalle IS 'gc_factura_recepcion_detalle: entidad del modulo gc; ver diccionario V4.';

COMMENT ON TABLE gv_pedido_factura IS 'gv_pedido_factura: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gl_asignacion_venta IS 'gl_asignacion_venta: entidad del modulo gl; ver diccionario V4.';

COMMENT ON TABLE gl_entrega_evento IS 'gl_entrega_evento: entidad del modulo gl; ver diccionario V4.';

COMMENT ON TABLE gi_movimiento IS 'gi_movimiento: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gi_incidencia IS 'gi_incidencia: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gi_reserva IS 'gi_reserva: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gv_devolucion IS 'gv_devolucion: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gc_devolucion IS 'gc_devolucion: entidad del modulo gc; ver diccionario V4.';

COMMENT ON TABLE gc_orden_compra_anulacion IS 'gc_orden_compra_anulacion: entidad del modulo gc; ver diccionario V4.';

COMMENT ON TABLE gc_recepcion_anulacion IS 'gc_recepcion_anulacion: entidad del modulo gc; ver diccionario V4.';

COMMENT ON TABLE gc_comprobante_anulacion IS 'gc_comprobante_anulacion: entidad del modulo gc; ver diccionario V4.';

COMMENT ON TABLE gv_cotizacion_anulacion IS 'gv_cotizacion_anulacion: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gv_pedido_anulacion IS 'gv_pedido_anulacion: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gv_comprobante_anulacion IS 'gv_comprobante_anulacion: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gi_ajuste_anulacion IS 'gi_ajuste_anulacion: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gi_transferencia_anulacion IS 'gi_transferencia_anulacion: entidad del modulo gi; ver diccionario V4.';

COMMENT ON TABLE gl_despacho_anulacion IS 'gl_despacho_anulacion: entidad del modulo gl; ver diccionario V4.';

COMMENT ON TABLE gv_nota_credito_anulacion IS 'gv_nota_credito_anulacion: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gc_nota_credito_anulacion IS 'gc_nota_credito_anulacion: entidad del modulo gc; ver diccionario V4.';

COMMENT ON TABLE gf_caja IS 'gf_caja: entidad del modulo gf; ver diccionario V4.';

COMMENT ON TABLE gf_caja_apertura IS 'gf_caja_apertura: entidad del modulo gf; ver diccionario V4.';

COMMENT ON TABLE gf_medio_pago IS 'gf_medio_pago: entidad del modulo gf; ver diccionario V4.';

COMMENT ON TABLE gv_cuota IS 'gv_cuota: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gc_cuota IS 'gc_cuota: entidad del modulo gc; ver diccionario V4.';

COMMENT ON TABLE gv_credito_aplicacion IS 'gv_credito_aplicacion: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gv_credito_disponible IS 'gv_credito_disponible: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gv_credito_uso IS 'gv_credito_uso: entidad del modulo gv; ver diccionario V4.';

COMMENT ON TABLE gc_credito_aplicacion IS 'gc_credito_aplicacion: entidad del modulo gc; ver diccionario V4.';

COMMENT ON TABLE gc_credito_disponible IS 'gc_credito_disponible: entidad del modulo gc; ver diccionario V4.';

COMMENT ON TABLE gc_credito_uso IS 'gc_credito_uso: entidad del modulo gc; ver diccionario V4.';

COMMENT ON TABLE gf_recibo_cabecera IS 'gf_recibo_cabecera: entidad del modulo gf; ver diccionario V4.';

COMMENT ON TABLE gf_recibo_forma IS 'gf_recibo_forma: entidad del modulo gf; ver diccionario V4.';

COMMENT ON TABLE gf_recibo_aplicacion IS 'gf_recibo_aplicacion: entidad del modulo gf; ver diccionario V4.';

COMMENT ON TABLE gf_pago_cabecera IS 'gf_pago_cabecera: entidad del modulo gf; ver diccionario V4.';

COMMENT ON TABLE gf_pago_forma IS 'gf_pago_forma: entidad del modulo gf; ver diccionario V4.';

COMMENT ON TABLE gf_pago_aplicacion IS 'gf_pago_aplicacion: entidad del modulo gf; ver diccionario V4.';

COMMENT ON TABLE gf_caja_arqueo IS 'gf_caja_arqueo: entidad del modulo gf; ver diccionario V4.';

COMMENT ON TABLE gf_cuenta_bancaria IS 'gf_cuenta_bancaria: entidad del modulo gf; ver diccionario V4.';

COMMENT ON TABLE gf_movimiento_bancario IS 'gf_movimiento_bancario: entidad del modulo gf; ver diccionario V4.';

COMMENT ON TABLE gp_permiso_offline IS 'gp_permiso_offline: entidad del modulo gp; ver diccionario V4.';

COMMENT ON TABLE gp_extension_permiso IS 'gp_extension_permiso: entidad del modulo gp; ver diccionario V4.';

COMMENT ON TABLE gp_evento_entrada IS 'gp_evento_entrada: entidad del modulo gp; ver diccionario V4.';

COMMENT ON TABLE gs_evento_salida IS 'gs_evento_salida: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE gs_auditoria IS 'gs_auditoria: entidad del modulo gs; ver diccionario V4.';

COMMENT ON TABLE fe_documento IS 'fe_documento: entidad del modulo fe; ver diccionario V4.';

COMMENT ON TABLE fe_evento IS 'fe_evento: entidad del modulo fe; ver diccionario V4.';

-- =====================================================================
-- 9. Reglas transaccionales e indices
-- =====================================================================

-- Contratos transaccionales. Funciones internas solo para el servicio confiable.
-- Las sesiones de usuarios finales NO reciben acceso SQL directo.
CREATE FUNCTION gs_acceso(e bigint,u bigint,s bigint) RETURNS void LANGUAGE plpgsql AS $$
BEGIN
 IF NOT EXISTS(SELECT 1 FROM gs_usuario_sucursal a JOIN gs_usuario x ON x.id=a.usuario_id
 JOIN gs_usuario_empresa z ON z.empresa_id=a.empresa_id AND z.usuario_id=a.usuario_id
 WHERE a.empresa_id=e AND a.usuario_id=u AND a.sucursal_id=s AND a.activo AND a.puede_emitir AND x.activo AND z.activo)
 THEN RAISE EXCEPTION 'USUARIO_NO_HABILITADO'; END IF;
END $$;

CREATE FUNCTION gs_validar_talonario(e bigint,t bigint,u bigint,f date) RETURNS void LANGUAGE plpgsql AS $$
DECLARE r gs_talonario; m gs_modalidad_emision;
BEGIN
 SELECT * INTO STRICT r FROM gs_talonario WHERE empresa_id=e AND id=t;
 PERFORM gs_acceso(e,u,r.sucursal_id);
 IF NOT r.activo OR f<r.desde OR (r.hasta IS NOT NULL AND f>r.hasta) THEN RAISE EXCEPTION 'TALONARIO_NO_VIGENTE'; END IF;
 IF NOT r.publico_sucursal AND NOT EXISTS(SELECT 1 FROM gs_talonario_usuario WHERE empresa_id=e AND talonario_id=t AND usuario_id=u)
 THEN RAISE EXCEPTION 'TALONARIO_RESTRINGIDO'; END IF;
 SELECT * INTO STRICT m FROM gs_modalidad_emision WHERE id=r.modalidad_id;
 IF m.fiscal AND (r.habilitacion_id IS NULL OR r.timbrado IS NULL OR r.establecimiento IS NULL OR r.punto_expedicion IS NULL
 OR r.final>9999999 OR NOT EXISTS(SELECT 1 FROM gs_habilitacion_fiscal h WHERE h.empresa_id=e AND h.id=r.habilitacion_id
 AND h.estado='VALIDADA' AND f>=h.desde AND (h.hasta IS NULL OR f<=h.hasta))) THEN RAISE EXCEPTION 'HABILITACION_FISCAL_INVALIDA'; END IF;
END $$;

CREATE FUNCTION gs_reservar_rango(e bigint,t bigint,terminal bigint,u bigint,n bigint,clave uuid,reserva boolean DEFAULT false)
RETURNS bigint LANGUAGE plpgsql AS $$
DECLARE r gs_talonario; x gs_talonario_rango_terminal; result bigint;
BEGIN
 PERFORM pg_advisory_xact_lock(e);
 SELECT * INTO x FROM gs_talonario_rango_terminal WHERE uid=clave;
 IF FOUND THEN
  IF x.empresa_id<>e OR x.talonario_id<>t OR x.terminal_id<>terminal OR x.fin-x.inicio+1<>n OR x.reserva_adicional<>reserva THEN RAISE EXCEPTION 'CLAVE_REUTILIZADA'; END IF;
  RETURN x.id;
 END IF;
 PERFORM gs_validar_talonario(e,t,u,current_date);
 SELECT * INTO STRICT r FROM gs_talonario WHERE empresa_id=e AND id=t FOR UPDATE;
 IF NOT EXISTS(SELECT 1 FROM gp_terminal WHERE empresa_id=e AND id=terminal AND sucursal_id=r.sucursal_id AND activa) THEN RAISE EXCEPTION 'TERMINAL_INVALIDA'; END IF;
 IF n<=0 OR n>r.final-r.siguiente+1 THEN RAISE EXCEPTION 'NUMERACION_AGOTADA'; END IF;
 INSERT INTO gs_talonario_rango_terminal(empresa_id,talonario_id,terminal_id,uid,inicio,fin,reserva_adicional)
 VALUES(e,t,terminal,clave,r.siguiente,r.siguiente+n-1,reserva) RETURNING id INTO result;
 UPDATE gs_talonario SET siguiente=siguiente+n WHERE id=t;
 RETURN result;
END $$;

CREATE FUNCTION gs_asignar_numero(e bigint,t bigint,u bigint,clave uuid,terminal bigint DEFAULT NULL,rango bigint DEFAULT NULL,
 n bigint DEFAULT NULL,fecha timestamptz DEFAULT now()) RETURNS bigint LANGUAGE plpgsql AS $$
DECLARE r gs_talonario; b gs_talonario_rango_terminal; x gs_talonario_numero; result bigint;
BEGIN
 PERFORM pg_advisory_xact_lock(e);
 SELECT * INTO x FROM gs_talonario_numero WHERE uid=clave;
 IF FOUND THEN
  IF x.empresa_id<>e OR x.talonario_id<>t OR x.usuario_id<>u OR x.terminal_id IS DISTINCT FROM terminal OR x.rango_id IS DISTINCT FROM rango
   OR (n IS NOT NULL AND x.numero<>n) THEN RAISE EXCEPTION 'CLAVE_REUTILIZADA'; END IF;
  RETURN x.id;
 END IF;
 PERFORM gs_validar_talonario(e,t,u,fecha::date);
 SELECT * INTO STRICT r FROM gs_talonario WHERE empresa_id=e AND id=t FOR UPDATE;
 IF terminal IS NOT NULL AND NOT EXISTS(SELECT 1 FROM gp_terminal WHERE empresa_id=e AND id=terminal AND sucursal_id=r.sucursal_id AND activa) THEN RAISE EXCEPTION 'TERMINAL_INVALIDA'; END IF;
 IF rango IS NULL THEN
  IF n IS NOT NULL THEN RAISE EXCEPTION 'NUMERO_MANUAL_PROHIBIDO'; END IF;
  IF r.siguiente>r.final THEN RAISE EXCEPTION 'NUMERACION_AGOTADA'; END IF;
  n:=r.siguiente; UPDATE gs_talonario SET siguiente=siguiente+1 WHERE id=t;
 ELSE
  SELECT * INTO STRICT b FROM gs_talonario_rango_terminal WHERE empresa_id=e AND id=rango AND talonario_id=t AND terminal_id=terminal;
  IF n IS NULL OR n<b.inicio OR n>b.fin THEN RAISE EXCEPTION 'FUERA_DE_RANGO'; END IF;
 END IF;
 INSERT INTO gs_talonario_numero(empresa_id,talonario_id,numero,uid,usuario_id,terminal_id,rango_id,clase_id,emitido_en)
 VALUES(e,t,n,clave,u,terminal,rango,r.clase_id,fecha) RETURNING id INTO result;
 RETURN result;
END $$;

CREATE FUNCTION gs_numero_disponible(e bigint,t bigint,u bigint,clave uuid,terminal bigint DEFAULT NULL) RETURNS bigint LANGUAGE plpgsql AS $$
DECLARE alt record; err text; existente gs_talonario_numero;
BEGIN
 SELECT * INTO existente FROM gs_talonario_numero WHERE uid=clave;
 IF FOUND THEN
  IF existente.empresa_id<>e OR existente.usuario_id<>u OR existente.terminal_id IS DISTINCT FROM terminal OR existente.rango_id IS NOT NULL
   OR (existente.talonario_id<>t AND NOT EXISTS(SELECT 1 FROM gs_talonario_alternativo WHERE empresa_id=e AND principal_id=t AND alternativo_id=existente.talonario_id)) THEN RAISE EXCEPTION 'CLAVE_REUTILIZADA'; END IF;
  RETURN existente.id;
 END IF;
 BEGIN RETURN gs_asignar_numero(e,t,u,clave,terminal);
 EXCEPTION WHEN OTHERS THEN GET STACKED DIAGNOSTICS err=MESSAGE_TEXT; IF err<>'NUMERACION_AGOTADA' THEN RAISE; END IF; END;
 FOR alt IN SELECT x.alternativo_id FROM gs_talonario_alternativo x JOIN gs_talonario p ON p.id=x.principal_id JOIN gs_talonario a ON a.id=x.alternativo_id
 WHERE x.empresa_id=e AND x.principal_id=t AND p.clase_id=a.clase_id AND p.sucursal_id=a.sucursal_id ORDER BY x.prioridad LOOP
  BEGIN RETURN gs_asignar_numero(e,alt.alternativo_id,u,clave,terminal);
  EXCEPTION WHEN OTHERS THEN GET STACKED DIAGNOSTICS err=MESSAGE_TEXT; IF err NOT IN ('NUMERACION_AGOTADA','TALONARIO_NO_VIGENTE','HABILITACION_FISCAL_INVALIDA','TALONARIO_RESTRINGIDO') THEN RAISE; END IF; END;
 END LOOP;
 RAISE EXCEPTION 'NUMERACION_AGOTADA: conservar operacion pendiente';
END $$;

CREATE FUNCTION gs_rango_guard() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE t gs_talonario;
BEGIN
 SELECT * INTO STRICT t FROM gs_talonario WHERE empresa_id=NEW.empresa_id AND id=NEW.talonario_id FOR UPDATE;
 IF NEW.inicio<t.inicial OR NEW.fin>t.final OR EXISTS(SELECT 1 FROM gs_talonario_rango_terminal WHERE empresa_id=NEW.empresa_id AND talonario_id=NEW.talonario_id AND inicio<=NEW.fin AND fin>=NEW.inicio)
 OR EXISTS(SELECT 1 FROM gs_talonario_numero WHERE empresa_id=NEW.empresa_id AND talonario_id=NEW.talonario_id AND numero BETWEEN NEW.inicio AND NEW.fin)
 THEN RAISE EXCEPTION 'RANGO_SUPERPUESTO_O_INVALIDO'; END IF;
 IF NOT EXISTS(SELECT 1 FROM gp_terminal WHERE empresa_id=NEW.empresa_id AND id=NEW.terminal_id AND sucursal_id=t.sucursal_id AND activa) THEN RAISE EXCEPTION 'TERMINAL_INVALIDA'; END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER rango_guard BEFORE INSERT ON gs_talonario_rango_terminal FOR EACH ROW EXECUTE FUNCTION gs_rango_guard();
CREATE FUNCTION gs_talonario_guard() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF NEW.siguiente<OLD.siguiente THEN RAISE EXCEPTION 'CONTADOR_NO_RETROCEDE'; END IF;
 IF EXISTS(SELECT 1 FROM gs_talonario_numero WHERE talonario_id=OLD.id) OR EXISTS(SELECT 1 FROM gs_talonario_rango_terminal WHERE talonario_id=OLD.id) THEN
  IF (to_jsonb(NEW)-ARRAY['siguiente','activo','publico_sucursal'])<>(to_jsonb(OLD)-ARRAY['siguiente','activo','publico_sucursal']) THEN RAISE EXCEPTION 'TALONARIO_UTILIZADO_INMUTABLE'; END IF;
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER talonario_guard BEFORE UPDATE ON gs_talonario FOR EACH ROW EXECUTE FUNCTION gs_talonario_guard();

CREATE FUNCTION gs_numero_exclusivo() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE t text; usos bigint; r gs_talonario_numero; cl bigint;
BEGIN
 IF NEW.numero_id IS NULL THEN RETURN NEW; END IF;
 SELECT * INTO STRICT r FROM gs_talonario_numero WHERE empresa_id=NEW.empresa_id AND id=NEW.numero_id FOR UPDATE;
 IF TG_TABLE_NAME IN ('gf_recibo_cabecera','gf_pago_cabecera') THEN
  SELECT id INTO cl FROM gs_clase_documento WHERE codigo=CASE WHEN TG_TABLE_NAME='gf_recibo_cabecera' THEN 'RECIBO' ELSE 'PAGO' END;
 ELSE
  SELECT clase_id INTO STRICT cl FROM gs_tipo_comprobante WHERE empresa_id=NEW.empresa_id AND id=NEW.tipo_id;
  IF NOT EXISTS(SELECT 1 FROM gs_talonario WHERE empresa_id=NEW.empresa_id AND id=r.talonario_id AND sucursal_id=NEW.sucursal_id)
    OR r.usuario_id<>NEW.usuario_id OR r.terminal_id IS DISTINCT FROM NEW.terminal_id THEN RAISE EXCEPTION 'NUMERO_CONTEXTO_INVALIDO'; END IF;
 END IF;
 IF r.clase_id<>cl THEN RAISE EXCEPTION 'CLASE_NUMERO_INVALIDA'; END IF;
 FOREACH t IN ARRAY ARRAY['gc_orden_compra_cabecera','gc_recepcion_cabecera','gc_comprobante_cabecera','gv_cotizacion_cabecera',
 'gv_pedido_cabecera','gv_comprobante_cabecera','gi_ajuste_cabecera','gi_transferencia_cabecera','gl_despacho_cabecera','gf_recibo_cabecera','gf_pago_cabecera','gv_nota_credito_cabecera','gc_nota_credito_cabecera'] LOOP
  EXECUTE format('SELECT count(*) FROM %I WHERE numero_id=$1 AND NOT ($2=$3 AND id=$4)',t) INTO usos USING NEW.numero_id,t,TG_TABLE_NAME,NEW.id;
  IF usos>0 THEN RAISE EXCEPTION 'NUMERO_YA_UTILIZADO'; END IF;
 END LOOP;
 RETURN NEW;
END $$;

CREATE FUNCTION gi_validar_posicion() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE a gi_articulo;
BEGIN
 SELECT * INTO STRICT a FROM gi_articulo WHERE empresa_id=NEW.empresa_id AND id=NEW.articulo_id;
 IF NOT a.mueve_stock OR a.controla_lote<>(NEW.lote_id IS NOT NULL) OR a.controla_serie<>(NEW.serie_id IS NOT NULL) THEN RAISE EXCEPTION 'TRAZABILIDAD_INVALIDA'; END IF;
 IF a.exige_vencimiento AND NOT EXISTS(SELECT 1 FROM gi_lote WHERE id=NEW.lote_id AND fecha_vencimiento IS NOT NULL) THEN RAISE EXCEPTION 'FALTA_VENCIMIENTO'; END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER validar_posicion BEFORE INSERT OR UPDATE ON gi_posicion FOR EACH ROW EXECUTE FUNCTION gi_validar_posicion();

CREATE FUNCTION gi_postear(e bigint,clave uuid,pos bigint,q numeric,costo numeric,permitir_negativo boolean,
 origen text,linea bigint,fecha timestamptz,motivo_texto text,valor_ajuste numeric DEFAULT 0,diferencia numeric DEFAULT 0)
RETURNS bigint LANGUAGE plpgsql AS $$
DECLARE a bigint; v gi_valoracion; s gi_existencia; ref numeric; aplicado numeric; delta numeric; gasto numeric:=diferencia;
 falta_costo boolean:=false; result bigint; seria bigint; serial_total numeric;
BEGIN
 PERFORM pg_advisory_xact_lock(e);
 IF EXISTS(SELECT 1 FROM gi_movimiento WHERE uid=clave) THEN RAISE EXCEPTION 'MOVIMIENTO_DUPLICADO'; END IF;
 SELECT articulo_id,serie_id INTO STRICT a,seria FROM gi_posicion WHERE empresa_id=e AND id=pos;
 INSERT INTO gi_valoracion(empresa_id,articulo_id) VALUES(e,a) ON CONFLICT DO NOTHING;
 INSERT INTO gi_existencia(empresa_id,posicion_id) VALUES(e,pos) ON CONFLICT DO NOTHING;
 SELECT * INTO STRICT v FROM gi_valoracion WHERE empresa_id=e AND articulo_id=a FOR UPDATE;
 SELECT * INTO STRICT s FROM gi_existencia WHERE empresa_id=e AND posicion_id=pos FOR UPDATE;
 IF q<0 AND NOT permitir_negativo AND s.cantidad-s.reservada+q<0 THEN RAISE EXCEPTION 'STOCK_INSUFICIENTE'; END IF;
 IF seria IS NOT NULL THEN
  SELECT coalesce(sum(x.cantidad),0) INTO serial_total FROM gi_existencia x JOIN gi_posicion p ON p.id=x.posicion_id WHERE p.empresa_id=e AND p.serie_id=seria;
  IF (q<>0 AND abs(q)<>1) OR s.cantidad+q NOT BETWEEN 0 AND 1 OR serial_total+q NOT BETWEEN 0 AND 1 THEN RAISE EXCEPTION 'SERIE_NO_DISPONIBLE'; END IF;
 END IF;
 SELECT coalesce(v.ultimo_costo,costo_referencia) INTO ref FROM gi_articulo WHERE empresa_id=e AND id=a;
 IF v.cantidad>0 AND NOT v.pendiente THEN ref:=v.valor/v.cantidad; END IF;
 aplicado:=CASE WHEN q>0 THEN costo ELSE coalesce(costo,ref) END;
 IF q=0 THEN delta:=valor_ajuste; aplicado:=NULL;
 ELSIF aplicado IS NULL THEN delta:=0; falta_costo:=true;
 ELSIF q>0 AND v.cantidad<0 THEN
  IF ref IS NULL OR v.pendiente THEN falta_costo:=true; delta:=q*aplicado;
  ELSE delta:=least(q,-v.cantidad)*ref + greatest(q+v.cantidad,0)*aplicado;
   gasto:=gasto+least(q,-v.cantidad)*(aplicado-ref);
  END IF;
 ELSE delta:=q*aplicado;
 END IF;
 IF q<0 AND v.cantidad+q=0 AND NOT v.pendiente AND costo IS NULL THEN delta:=-v.valor; END IF;
 IF q<0 AND v.cantidad+q=0 AND NOT v.pendiente AND costo IS NOT NULL THEN gasto:=gasto+v.valor+delta; delta:=-v.valor; END IF;
 IF origen='REVERSION' THEN delta:=valor_ajuste; gasto:=diferencia; END IF;
 IF v.cantidad+q>0 AND v.valor+delta<0 AND NOT v.pendiente THEN RAISE EXCEPTION 'VALOR_INVENTARIO_NEGATIVO'; END IF;
 IF v.cantidad+q=0 AND abs(v.valor+delta)>0.000001 AND NOT v.pendiente THEN gasto:=gasto+v.valor+delta;delta:=-v.valor; END IF;
 UPDATE gi_existencia SET cantidad=cantidad+q WHERE id=s.id;
 UPDATE gi_valoracion SET cantidad=cantidad+q,valor=valor+delta,pendiente=pendiente OR falta_costo,
  ultimo_costo=CASE WHEN cantidad+q>0 AND NOT (pendiente OR falta_costo) THEN (valor+delta)/(cantidad+q) ELSE coalesce(ultimo_costo,aplicado) END,
  version=version+1 WHERE id=v.id;
 INSERT INTO gi_movimiento(empresa_id,uid,articulo_id,posicion_id,fecha_operacion,cantidad,valor_inventario,diferencia_consumida,costo_unitario,
 cantidad_antes,cantidad_despues,valor_antes,valor_despues,valoracion_pendiente,motivo,
 recepcion_detalle_id,venta_detalle_id,ajuste_detalle_id,transferencia_detalle_id,factura_recepcion_id,reversion_de)
 VALUES(e,clave,a,pos,fecha,q,delta,gasto,aplicado,v.cantidad,v.cantidad+q,v.valor,v.valor+delta,v.pendiente OR falta_costo,motivo_texto,
 CASE WHEN origen='RECEPCION' THEN linea END,CASE WHEN origen='VENTA' THEN linea END,CASE WHEN origen='AJUSTE' THEN linea END,
 CASE WHEN origen='TRANSFERENCIA' THEN linea END,CASE WHEN origen='FACTURA' THEN linea END,CASE WHEN origen='REVERSION' THEN linea END) RETURNING id INTO result;
 IF s.cantidad+q<0 THEN INSERT INTO gi_incidencia(empresa_id,movimiento_id,tipo) VALUES(e,result,'NEGATIVO'); END IF;
 IF falta_costo THEN INSERT INTO gi_incidencia(empresa_id,movimiento_id,tipo) VALUES(e,result,'SIN_COSTO'); END IF;
 RETURN result;
END $$;

CREATE FUNCTION gs_inmutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'REGISTRO_INMUTABLE: %',TG_TABLE_NAME; END $$;
CREATE TRIGGER movimiento_inmutable BEFORE UPDATE OR DELETE ON gi_movimiento FOR EACH ROW EXECUTE FUNCTION gs_inmutable();
CREATE TRIGGER rango_inmutable BEFORE UPDATE OR DELETE ON gs_talonario_rango_terminal FOR EACH ROW EXECUTE FUNCTION gs_inmutable();

CREATE FUNCTION gs_validar_numero_edicion() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF TG_OP='DELETE' OR (to_jsonb(NEW)-'estado')<>(to_jsonb(OLD)-'estado') OR OLD.estado='ANULADO' OR NEW.estado<>'ANULADO' THEN RAISE EXCEPTION 'NUMERO_INMUTABLE'; END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER numero_inmutable BEFORE UPDATE OR DELETE ON gs_talonario_numero FOR EACH ROW EXECUTE FUNCTION gs_validar_numero_edicion();

CREATE FUNCTION gc_aplicar_diferencia(e bigint,vinculo bigint,proporcion numeric DEFAULT NULL) RETURNS void LANGUAGE plpgsql AS $$
DECLARE x gc_factura_recepcion_detalle; r gc_recepcion_detalle; f gc_comprobante_detalle; h gc_recepcion_cabecera;
 fh gc_comprobante_cabecera; q numeric; delta numeric; cap numeric; cant numeric;
BEGIN
 PERFORM pg_advisory_xact_lock(e);
 SELECT * INTO STRICT x FROM gc_factura_recepcion_detalle WHERE empresa_id=e AND id=vinculo FOR UPDATE;
 IF x.aplicado THEN RETURN; END IF;
 SELECT * INTO STRICT r FROM gc_recepcion_detalle WHERE empresa_id=e AND id=x.recepcion_detalle_id;
 SELECT * INTO STRICT f FROM gc_comprobante_detalle WHERE empresa_id=e AND id=x.factura_detalle_id;
 SELECT * INTO STRICT h FROM gc_recepcion_cabecera WHERE id=r.cabecera_id;
 SELECT * INTO STRICT fh FROM gc_comprobante_cabecera WHERE id=f.cabecera_id;
 IF h.estado<>'CONFIRMADO' OR h.tercero_id<>fh.tercero_id OR r.articulo_id<>f.articulo_id THEN RAISE EXCEPTION 'RECEPCION_INCOMPATIBLE'; END IF;
 SELECT coalesce(sum(x2.cantidad_base),0) INTO q FROM gc_factura_recepcion_detalle x2
 JOIN gc_comprobante_detalle fd ON fd.id=x2.factura_detalle_id JOIN gc_comprobante_cabecera fc ON fc.id=fd.cabecera_id
 WHERE x2.empresa_id=e AND x2.recepcion_detalle_id=r.id AND fc.estado<>'REVERTIDO';
 IF q>r.cantidad_base THEN RAISE EXCEPTION 'RECEPCION_SOBREFACTURADA'; END IF;
 SELECT coalesce(sum(cantidad_base),0) INTO q FROM gc_factura_recepcion_detalle WHERE empresa_id=e AND factura_detalle_id=f.id;
 IF q>f.cantidad_base THEN RAISE EXCEPTION 'FACTURA_SOBREAPLICADA'; END IF;
 delta:=(x.costo_final_base-r.costo_unitario_base)*x.cantidad_base;
 SELECT greatest(cantidad,0) INTO cant FROM gi_valoracion WHERE empresa_id=e AND articulo_id=r.articulo_id;
 -- Politica de cobertura actual: no pretende rastrear unidades fisicas por factura en promedio movil.
 SELECT sum(x2.cantidad_base) INTO q FROM gc_factura_recepcion_detalle x2 JOIN gc_comprobante_detalle d2 ON d2.id=x2.factura_detalle_id
 WHERE x2.empresa_id=e AND d2.cabecera_id=fh.id AND d2.articulo_id=r.articulo_id;
 IF proporcion IS NOT NULL AND (proporcion<0 OR proporcion>1) THEN RAISE EXCEPTION 'PROPORCION_INVALIDA'; END IF;
 cap:=delta*coalesce(proporcion,least(coalesce(cant,0),q)/q);
 IF delta<>0 THEN PERFORM gi_postear(e,gen_random_uuid(),r.posicion_id,0,NULL,false,'FACTURA',x.id,fh.fecha_operacion,'DIFERENCIA_FACTURA',cap,delta-cap); END IF;
 UPDATE gc_factura_recepcion_detalle SET aplicado=true,diferencia_base=delta WHERE id=x.id;
END $$;

CREATE FUNCTION gc_vinculo_guard() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE h bigint; estado text;
BEGIN
 IF TG_OP<>'INSERT' AND OLD.aplicado THEN RAISE EXCEPTION 'VINCULO_APLICADO_INMUTABLE'; END IF;
 IF TG_OP='UPDATE' AND (to_jsonb(NEW)-ARRAY['aplicado','diferencia_base'])<>(to_jsonb(OLD)-ARRAY['aplicado','diferencia_base']) THEN RAISE EXCEPTION 'RECREAR_VINCULO_BORRADOR'; END IF;
 IF TG_OP='DELETE' THEN RETURN OLD; END IF;
 IF TG_OP='INSERT' AND (NEW.aplicado OR NEW.diferencia_base IS NOT NULL) THEN RAISE EXCEPTION 'VINCULO_NO_APLICADO'; END IF;
 SELECT c.id,c.estado INTO STRICT h,estado FROM gc_comprobante_detalle d JOIN gc_comprobante_cabecera c ON c.id=d.cabecera_id WHERE d.empresa_id=NEW.empresa_id AND d.id=NEW.factura_detalle_id FOR UPDATE OF c;
 IF estado='REVERTIDO' THEN RAISE EXCEPTION 'FACTURA_REVERTIDA'; END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER vinculo_guard BEFORE INSERT OR UPDATE OR DELETE ON gc_factura_recepcion_detalle FOR EACH ROW EXECUTE FUNCTION gc_vinculo_guard();
CREATE FUNCTION gc_aplicar_vinculos_posteriores(e bigint,factura bigint) RETURNS void LANGUAGE plpgsql AS $$
DECLARE grupo record; x record; disponible numeric; proporcion numeric;
BEGIN
 PERFORM pg_advisory_xact_lock(e);
 IF NOT EXISTS(SELECT 1 FROM gc_comprobante_cabecera WHERE empresa_id=e AND id=factura AND estado='CONFIRMADO') THEN RAISE EXCEPTION 'FACTURA_NO_CONFIRMADA'; END IF;
 FOR grupo IN SELECT d.articulo_id,sum(v.cantidad_base) cantidad FROM gc_factura_recepcion_detalle v JOIN gc_comprobante_detalle d ON d.id=v.factura_detalle_id WHERE v.empresa_id=e AND d.cabecera_id=factura AND NOT v.aplicado GROUP BY d.articulo_id LOOP
  SELECT greatest(cantidad,0) INTO disponible FROM gi_valoracion WHERE empresa_id=e AND articulo_id=grupo.articulo_id;
  proporcion:=least(coalesce(disponible,0),grupo.cantidad)/grupo.cantidad;
  FOR x IN SELECT v.id FROM gc_factura_recepcion_detalle v JOIN gc_comprobante_detalle d ON d.id=v.factura_detalle_id WHERE v.empresa_id=e AND d.cabecera_id=factura AND d.articulo_id=grupo.articulo_id AND NOT v.aplicado ORDER BY v.id LOOP
   PERFORM gc_aplicar_diferencia(e,x.id,proporcion);
  END LOOP;
 END LOOP;
END $$;

CREATE FUNCTION gs_vinculo_cantidades() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE iz record; de record; usado numeric; col1 text:=TG_ARGV[0]; col2 text:=TG_ARGV[1]; tabla1 text:=TG_ARGV[2]; tabla2 text:=TG_ARGV[3];
 id1 bigint; id2 bigint; tercero1 bigint; tercero2 bigint; est text;
BEGIN
 IF TG_OP<>'INSERT' THEN RAISE EXCEPTION 'VINCULO_INMUTABLE'; END IF;
 PERFORM pg_advisory_xact_lock(NEW.empresa_id);
 id1:=(to_jsonb(NEW)->>col1)::bigint; id2:=(to_jsonb(NEW)->>col2)::bigint;
 EXECUTE format('SELECT * FROM %I WHERE empresa_id=$1 AND id=$2',tabla1) INTO STRICT iz USING NEW.empresa_id,id1;
 EXECUTE format('SELECT * FROM %I WHERE empresa_id=$1 AND id=$2',tabla2) INTO STRICT de USING NEW.empresa_id,id2;
 IF iz.articulo_id<>de.articulo_id THEN RAISE EXCEPTION 'ARTICULOS_DIFERENTES'; END IF;
 EXECUTE format('SELECT estado FROM %I WHERE empresa_id=$1 AND id=$2',replace(tabla1,'_detalle','_cabecera')) INTO est USING NEW.empresa_id,iz.cabecera_id;
 IF est<>'BORRADOR' THEN RAISE EXCEPTION 'VINCULAR_ANTES_DE_CONFIRMAR'; END IF;
 EXECUTE format('SELECT coalesce(sum(cantidad_base),0) FROM %I WHERE empresa_id=$1 AND %I=$2',TG_TABLE_NAME,col2) INTO usado USING NEW.empresa_id,id2;
 IF usado+NEW.cantidad_base>de.cantidad_base THEN RAISE EXCEPTION 'CANTIDAD_ORIGEN_EXCEDIDA'; END IF;
 EXECUTE format('SELECT coalesce(sum(cantidad_base),0) FROM %I WHERE empresa_id=$1 AND %I=$2',TG_TABLE_NAME,col1) INTO usado USING NEW.empresa_id,id1;
 IF usado+NEW.cantidad_base>iz.cantidad_base THEN RAISE EXCEPTION 'CANTIDAD_DESTINO_EXCEDIDA'; END IF;
 IF TG_TABLE_NAME='gc_recepcion_orden' THEN
  SELECT tercero_id INTO tercero1 FROM gc_recepcion_cabecera WHERE id=iz.cabecera_id;
  SELECT tercero_id INTO tercero2 FROM gc_orden_compra_cabecera WHERE id=de.cabecera_id;
  IF tercero1<>tercero2 THEN RAISE EXCEPTION 'PROVEEDOR_DIFERENTE'; END IF;
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER recepcion_orden_validar BEFORE INSERT OR UPDATE OR DELETE ON gc_recepcion_orden FOR EACH ROW EXECUTE FUNCTION gs_vinculo_cantidades('recepcion_detalle_id','orden_detalle_id','gc_recepcion_detalle','gc_orden_compra_detalle');
CREATE TRIGGER despacho_validar BEFORE INSERT OR UPDATE OR DELETE ON gl_asignacion_venta FOR EACH ROW EXECUTE FUNCTION gs_vinculo_cantidades('despacho_detalle_id','venta_detalle_id','gl_despacho_detalle','gv_comprobante_detalle');
CREATE TRIGGER pedido_validar BEFORE INSERT OR UPDATE OR DELETE ON gv_pedido_factura FOR EACH ROW EXECUTE FUNCTION gs_vinculo_cantidades('factura_detalle_id','pedido_detalle_id','gv_comprobante_detalle','gv_pedido_detalle');

CREATE FUNCTION gs_confirmacion() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE d record; t gs_tipo_comprobante; cl text; total_lineas numeric; signo integer; es_pos boolean:=false; coste numeric; v numeric;
 origen record; restante numeric; otros numeric; cu record; ap numeric; pref text;
BEGIN
 IF TG_OP='DELETE' THEN IF OLD.estado<>'BORRADOR' THEN RAISE EXCEPTION 'DOCUMENTO_INMUTABLE'; END IF; RETURN OLD; END IF;
 IF TG_OP='INSERT' AND NEW.estado<>'BORRADOR' THEN RAISE EXCEPTION 'CREAR_COMO_BORRADOR'; END IF;
 IF TG_OP='UPDATE' AND OLD.estado<>'BORRADOR' THEN
  IF NEW.estado='REVERTIDO' AND OLD.estado='CONFIRMADO' AND (to_jsonb(NEW)-'estado')=(to_jsonb(OLD)-'estado') THEN
   EXECUTE format('SELECT count(*) FROM %I WHERE empresa_id=$1 AND cabecera_id=$2',replace(TG_TABLE_NAME,'_cabecera','_anulacion')) INTO total_lineas USING NEW.empresa_id,NEW.id;
   IF total_lineas=1 THEN RETURN NEW; END IF;
  END IF;
  RAISE EXCEPTION 'DOCUMENTO_INMUTABLE';
 END IF;
 SELECT * INTO STRICT t FROM gs_tipo_comprobante WHERE empresa_id=NEW.empresa_id AND id=NEW.tipo_id;
 SELECT codigo INTO STRICT cl FROM gs_clase_documento WHERE id=t.clase_id;
 IF cl<>(CASE TG_TABLE_NAME WHEN 'gc_orden_compra_cabecera' THEN 'ORDEN_COMPRA' WHEN 'gc_recepcion_cabecera' THEN 'RECEPCION'
 WHEN 'gc_comprobante_cabecera' THEN 'COMPRA' WHEN 'gv_cotizacion_cabecera' THEN 'COTIZACION' WHEN 'gv_pedido_cabecera' THEN 'PEDIDO'
 WHEN 'gv_comprobante_cabecera' THEN 'VENTA' WHEN 'gi_ajuste_cabecera' THEN 'AJUSTE' WHEN 'gi_transferencia_cabecera' THEN 'TRANSFERENCIA'
 WHEN 'gl_despacho_cabecera' THEN 'DESPACHO' WHEN 'gv_nota_credito_cabecera' THEN 'CREDITO_VENTA' WHEN 'gc_nota_credito_cabecera' THEN 'CREDITO_COMPRA' END) THEN RAISE EXCEPTION 'TIPO_MODULO_INCORRECTO'; END IF;
 IF NEW.estado<>'CONFIRMADO' THEN RETURN NEW; END IF;
 PERFORM pg_advisory_xact_lock(NEW.empresa_id);
 PERFORM gs_acceso(NEW.empresa_id,NEW.usuario_id,NEW.sucursal_id);
 IF NOT t.activo THEN RAISE EXCEPTION 'TIPO_INACTIVO'; END IF;
 IF TG_TABLE_NAME<>'gc_comprobante_cabecera' AND NEW.numero_id IS NULL THEN RAISE EXCEPTION 'FALTA_NUMERO'; END IF;
 NEW.efecto_stock_aplicado:=coalesce(NEW.efecto_stock_aplicado,t.efecto_stock_predeterminado);
 IF NEW.efecto_stock_aplicado<>t.efecto_stock_predeterminado AND (NOT t.permite_excepcion OR NEW.motivo_excepcion IS NULL OR
 NOT EXISTS(SELECT 1 FROM gs_usuario_sucursal WHERE empresa_id=NEW.empresa_id AND usuario_id=NEW.usuario_id AND sucursal_id=NEW.sucursal_id AND supervisor)) THEN RAISE EXCEPTION 'EXCEPCION_NO_AUTORIZADA'; END IF;
 IF TG_TABLE_NAME IN ('gc_comprobante_cabecera','gc_orden_compra_cabecera','gv_pedido_cabecera','gv_cotizacion_cabecera','gl_despacho_cabecera','gv_nota_credito_cabecera','gc_nota_credito_cabecera') AND NEW.efecto_stock_aplicado<>'NINGUNO' THEN RAISE EXCEPTION 'CIRCUITO_SIN_MOVIMIENTO'; END IF;
 IF TG_TABLE_NAME='gc_recepcion_cabecera' AND NEW.efecto_stock_aplicado<>'ENTRADA' THEN RAISE EXCEPTION 'RECEPCION_REQUIERE_ENTRADA'; END IF;
 IF TG_TABLE_NAME='gv_comprobante_cabecera' THEN
  es_pos:=NEW.origen='POS';
  IF es_pos AND NOT EXISTS(SELECT 1 FROM gp_terminal WHERE empresa_id=NEW.empresa_id AND id=NEW.terminal_id AND sucursal_id=NEW.sucursal_id AND activa) THEN RAISE EXCEPTION 'POS_REQUIERE_TERMINAL'; END IF;
  IF NEW.efecto_stock_aplicado='ENTRADA' THEN RAISE EXCEPTION 'UTILIZAR_DEVOLUCION_VINCULADA'; END IF;
 END IF;
 EXECUTE format('SELECT sum(importe_total) FROM %I WHERE empresa_id=$1 AND cabecera_id=$2',replace(TG_TABLE_NAME,'_cabecera','_detalle')) INTO total_lineas USING NEW.empresa_id,NEW.id;
 IF total_lineas IS NULL THEN RAISE EXCEPTION 'DOCUMENTO_SIN_LINEAS'; END IF;
 NEW.total:=total_lineas;
 FOR d IN EXECUTE format('SELECT l.*,a.mueve_stock FROM %I l JOIN gi_articulo a ON a.id=l.articulo_id WHERE l.empresa_id=$1 AND l.cabecera_id=$2 ORDER BY l.articulo_id,l.id',replace(TG_TABLE_NAME,'_cabecera','_detalle')) USING NEW.empresa_id,NEW.id LOOP
  IF d.impuesto<>round(d.base_imponible*d.tasa_impuesto/100,6) OR d.base_imponible>d.cantidad*d.precio-d.descuento THEN RAISE EXCEPTION 'IMPUESTO_INCONSISTENTE'; END IF;
  IF d.mueve_stock AND NEW.efecto_stock_aplicado<>'NINGUNO' THEN
   IF d.posicion_id IS NULL THEN RAISE EXCEPTION 'FALTA_POSICION'; END IF;
   IF NOT EXISTS(SELECT 1 FROM gi_posicion p JOIN gi_ubicacion u ON u.id=p.ubicacion_id JOIN gi_deposito b ON b.id=u.deposito_id WHERE p.id=d.posicion_id AND b.sucursal_id=NEW.sucursal_id) THEN RAISE EXCEPTION 'DEPOSITO_OTRA_SUCURSAL'; END IF;
   signo:=CASE WHEN NEW.efecto_stock_aplicado='ENTRADA' THEN 1 ELSE -1 END;
   IF TG_TABLE_NAME='gc_recepcion_cabecera' THEN
    IF d.costo_unitario_base IS NULL THEN RAISE EXCEPTION 'RECEPCION_REQUIERE_COSTO'; END IF;
    PERFORM gi_postear(NEW.empresa_id,gen_random_uuid(),d.posicion_id,d.cantidad_base,d.costo_unitario_base,false,'RECEPCION',d.id,NEW.fecha_operacion,'RECEPCION');
   ELSIF TG_TABLE_NAME='gv_comprobante_cabecera' THEN
    PERFORM gi_postear(NEW.empresa_id,gen_random_uuid(),d.posicion_id,-d.cantidad_base,NULL,es_pos,'VENTA',d.id,NEW.fecha_operacion,'VENTA');
   ELSIF TG_TABLE_NAME='gi_ajuste_cabecera' THEN
    IF NEW.motivo_excepcion IS NULL THEN RAISE EXCEPTION 'AJUSTE_REQUIERE_MOTIVO'; END IF;
    IF signo=1 AND d.costo_unitario_base IS NULL THEN RAISE EXCEPTION 'ENTRADA_REQUIERE_COSTO'; END IF;
    PERFORM gi_postear(NEW.empresa_id,gen_random_uuid(),d.posicion_id,signo*d.cantidad_base,CASE WHEN signo=1 THEN d.costo_unitario_base END,false,'AJUSTE',d.id,NEW.fecha_operacion,'AJUSTE');
   ELSIF TG_TABLE_NAME='gi_transferencia_cabecera' THEN
    IF d.destino_id IS NULL OR d.destino_id=d.posicion_id THEN RAISE EXCEPTION 'DESTINO_INVALIDO'; END IF;
    IF NOT EXISTS(SELECT 1 FROM gi_posicion p JOIN gi_posicion z ON z.id=d.destino_id WHERE p.id=d.posicion_id AND p.lote_id IS NOT DISTINCT FROM z.lote_id AND p.serie_id IS NOT DISTINCT FROM z.serie_id) THEN RAISE EXCEPTION 'TRANSFERENCIA_CAMBIA_TRAZABILIDAD'; END IF;
    PERFORM gi_postear(NEW.empresa_id,gen_random_uuid(),d.posicion_id,-d.cantidad_base,NULL,false,'TRANSFERENCIA',d.id,NEW.fecha_operacion,'TRANSFERENCIA_SALIDA');
    SELECT costo_unitario INTO coste FROM gi_movimiento WHERE transferencia_detalle_id=d.id AND motivo='TRANSFERENCIA_SALIDA';
    IF coste IS NULL THEN RAISE EXCEPTION 'TRANSFERENCIA_SIN_COSTO'; END IF;
    PERFORM gi_postear(NEW.empresa_id,gen_random_uuid(),d.destino_id,d.cantidad_base,coste,false,'TRANSFERENCIA',d.id,NEW.fecha_operacion,'TRANSFERENCIA_ENTRADA');
   END IF;
  END IF;
  IF TG_TABLE_NAME='gc_comprobante_cabecera' THEN
   FOR v IN SELECT id FROM gc_factura_recepcion_detalle WHERE empresa_id=NEW.empresa_id AND factura_detalle_id=d.id ORDER BY id LOOP
    PERFORM gc_aplicar_diferencia(NEW.empresa_id,v::bigint);
   END LOOP;
  END IF;
 END LOOP;
 IF TG_TABLE_NAME IN ('gv_comprobante_cabecera','gc_comprobante_cabecera') AND t.efecto_financiero=1 AND NEW.total>0 THEN
  EXECUTE format('INSERT INTO %I(empresa_id,comprobante_id,numero,vencimiento,importe) VALUES($1,$2,1,$3,$4)',left(TG_TABLE_NAME,2)||'_cuota') USING NEW.empresa_id,NEW.id,NEW.fecha_operacion::date,NEW.total;
 END IF;
 IF TG_TABLE_NAME IN ('gv_nota_credito_cabecera','gc_nota_credito_cabecera') THEN
  IF t.efecto_financiero<>-1 THEN RAISE EXCEPTION 'NOTA_REQUIERE_EFECTO_CREDITO'; END IF;
  pref:=left(TG_TABLE_NAME,2);
  EXECUTE format('SELECT * FROM %I WHERE empresa_id=$1 AND id=$2 FOR UPDATE',pref||'_comprobante_cabecera') INTO STRICT origen USING NEW.empresa_id,NEW.comprobante_origen_id;
  IF origen.estado<>'CONFIRMADO' OR origen.tercero_id<>NEW.tercero_id OR origen.moneda_id<>NEW.moneda_id THEN RAISE EXCEPTION 'ORIGEN_CREDITO_INVALIDO'; END IF;
  EXECUTE format('SELECT coalesce(sum(total),0) FROM %I WHERE empresa_id=$1 AND comprobante_origen_id=$2 AND estado=''CONFIRMADO'' AND id<>$3',TG_TABLE_NAME) INTO otros USING NEW.empresa_id,NEW.comprobante_origen_id,NEW.id;
  IF NEW.total+otros>origen.total THEN RAISE EXCEPTION 'CREDITO_EXCEDE_ORIGEN'; END IF;
  restante:=NEW.total;
  FOR cu IN EXECUTE format('SELECT * FROM %I WHERE empresa_id=$1 AND comprobante_id=$2 ORDER BY numero FOR UPDATE',pref||'_cuota') USING NEW.empresa_id,origen.id LOOP
   ap:=least(restante,cu.saldo);
   IF ap>0 THEN
    EXECUTE format('INSERT INTO %I(empresa_id,nota_id,cuota_id,importe) VALUES($1,$2,$3,$4)',pref||'_credito_aplicacion') USING NEW.empresa_id,NEW.id,cu.id,ap;
    EXECUTE format('UPDATE %I SET creditado=creditado+$1 WHERE id=$2',pref||'_cuota') USING ap,cu.id;
    restante:=restante-ap;
   END IF;
  END LOOP;
  IF restante>0 THEN EXECUTE format('INSERT INTO %I(empresa_id,nota_id,importe) VALUES($1,$2,$3)',pref||'_credito_disponible') USING NEW.empresa_id,NEW.id,restante; END IF;
 ELSIF t.efecto_financiero=-1 THEN RAISE EXCEPTION 'USAR_NOTA_CREDITO'; END IF;
 NEW.confirmado_en:=clock_timestamp();
 INSERT INTO gs_auditoria(empresa_id,usuario_id,tabla,registro_id,accion,despues) VALUES(NEW.empresa_id,NEW.usuario_id,TG_TABLE_NAME,NEW.id,'CONFIRMAR',to_jsonb(NEW));
 INSERT INTO gs_evento_salida(empresa_id,tema,contenido) VALUES(NEW.empresa_id,TG_TABLE_NAME||'.confirmado',jsonb_build_object('id',NEW.id,'uid',NEW.uid));
 RETURN NEW;
END $$;

CREATE FUNCTION gs_linea_borrador() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE estado text; cab bigint; emp bigint;
BEGIN
 cab:=CASE WHEN TG_OP='DELETE' THEN OLD.cabecera_id ELSE NEW.cabecera_id END;
 emp:=CASE WHEN TG_OP='DELETE' THEN OLD.empresa_id ELSE NEW.empresa_id END;
 IF TG_OP='UPDATE' AND (NEW.cabecera_id<>OLD.cabecera_id OR NEW.empresa_id<>OLD.empresa_id) THEN RAISE EXCEPTION 'LINEA_NO_TRANSFERIBLE'; END IF;
 EXECUTE format('SELECT estado FROM %I WHERE empresa_id=$1 AND id=$2 FOR UPDATE',replace(TG_TABLE_NAME,'_detalle','_cabecera')) INTO estado USING emp,cab;
 IF estado IS DISTINCT FROM 'BORRADOR' THEN RAISE EXCEPTION 'DETALLE_INMUTABLE'; END IF;
 IF TG_OP='DELETE' THEN RETURN OLD; END IF;
 RETURN NEW;
END $$;

DO $$ DECLARE b text; BEGIN
 FOREACH b IN ARRAY ARRAY['gc_orden_compra','gc_recepcion','gc_comprobante','gv_cotizacion','gv_pedido','gv_comprobante','gi_ajuste','gi_transferencia','gl_despacho','gv_nota_credito','gc_nota_credito'] LOOP
  EXECUTE format('CREATE TRIGGER a_confirmar BEFORE INSERT OR UPDATE OR DELETE ON %I FOR EACH ROW EXECUTE FUNCTION gs_confirmacion()',b||'_cabecera');
  EXECUTE format('CREATE TRIGGER b_numero BEFORE INSERT OR UPDATE ON %I FOR EACH ROW EXECUTE FUNCTION gs_numero_exclusivo()',b||'_cabecera');
  EXECUTE format('CREATE TRIGGER linea_borrador BEFORE INSERT OR UPDATE OR DELETE ON %I FOR EACH ROW EXECUTE FUNCTION gs_linea_borrador()',b||'_detalle');
 END LOOP;
END $$;

CREATE FUNCTION gs_confirmar(e bigint,modulo text,documento bigint) RETURNS void LANGUAGE plpgsql AS $$
DECLARE estado text;
BEGIN
 IF modulo NOT IN ('gc_orden_compra','gc_recepcion','gc_comprobante','gv_cotizacion','gv_pedido','gv_comprobante','gi_ajuste','gi_transferencia','gl_despacho','gv_nota_credito','gc_nota_credito') THEN RAISE EXCEPTION 'MODULO_INVALIDO'; END IF;
 PERFORM pg_advisory_xact_lock(e);
 EXECUTE format('SELECT estado FROM %I WHERE empresa_id=$1 AND id=$2 FOR UPDATE',modulo||'_cabecera') INTO STRICT estado USING e,documento;
 IF estado='CONFIRMADO' THEN RETURN; END IF;
 EXECUTE format('UPDATE %I SET estado=''CONFIRMADO'' WHERE empresa_id=$1 AND id=$2',modulo||'_cabecera') USING e,documento;
END $$;

CREATE FUNCTION gf_aplicar(e bigint,clase text,documento bigint) RETURNS void LANGUAGE plpgsql AS $$
DECLARE h record; a record; c record; n numeric; total_aplicado numeric; cuota_tabla text; tercero bigint; moneda bigint;
BEGIN
 IF clase NOT IN ('gf_recibo','gf_pago') THEN RAISE EXCEPTION 'CLASE_INVALIDA'; END IF;
 PERFORM pg_advisory_xact_lock(e);
 EXECUTE format('SELECT * FROM %I WHERE empresa_id=$1 AND id=$2 FOR UPDATE',clase||'_cabecera') INTO STRICT h USING e,documento;
 IF h.estado='CONFIRMADO' THEN RETURN; END IF;
 IF h.estado<>'BORRADOR' THEN RAISE EXCEPTION 'ESTADO_INVALIDO'; END IF;
 IF h.numero_id IS NULL THEN RAISE EXCEPTION 'FALTA_NUMERO'; END IF;
 IF h.apertura_id IS NOT NULL AND NOT EXISTS(SELECT 1 FROM gf_caja_apertura ap JOIN gf_caja ca ON ca.id=ap.caja_id WHERE ap.empresa_id=e AND ap.id=h.apertura_id AND ap.cerrada_en IS NULL AND ca.moneda_id=h.moneda_id) THEN RAISE EXCEPTION 'CAJA_CERRADA_O_MONEDA'; END IF;
 EXECUTE format('SELECT coalesce(sum(importe),0) FROM %I WHERE empresa_id=$1 AND cabecera_id=$2',clase||'_forma') INTO n USING e,documento;
 IF n<>h.importe THEN RAISE EXCEPTION 'MEDIOS_NO_CUADRAN'; END IF;
 EXECUTE format('SELECT coalesce(sum(importe),0) FROM %I WHERE empresa_id=$1 AND cabecera_id=$2',clase||'_aplicacion') INTO total_aplicado USING e,documento;
 IF total_aplicado>h.importe THEN RAISE EXCEPTION 'APLICACION_EXCEDE_IMPORTE'; END IF;
 cuota_tabla:=CASE WHEN clase='gf_recibo' THEN 'gv_cuota' ELSE 'gc_cuota' END;
 FOR a IN EXECUTE format('SELECT * FROM %I WHERE empresa_id=$1 AND cabecera_id=$2 ORDER BY cuota_id',clase||'_aplicacion') USING e,documento LOOP
  EXECUTE format('SELECT * FROM %I WHERE empresa_id=$1 AND id=$2 FOR UPDATE',cuota_tabla) INTO STRICT c USING e,a.cuota_id;
  EXECUTE format('SELECT tercero_id,moneda_id FROM %I WHERE empresa_id=$1 AND id=$2 AND estado=''CONFIRMADO''',left(cuota_tabla,2)||'_comprobante_cabecera') INTO tercero,moneda USING e,c.comprobante_id;
  IF tercero IS NULL THEN RAISE EXCEPTION 'COMPROBANTE_NO_CONFIRMADO'; END IF;
  IF tercero<>h.tercero_id OR moneda<>h.moneda_id THEN RAISE EXCEPTION 'CUOTA_TERCERO_MONEDA'; END IF;
  IF a.importe>c.saldo THEN RAISE EXCEPTION 'PAGO_EXCEDE_SALDO'; END IF;
  EXECUTE format('UPDATE %I SET aplicado=aplicado+$1 WHERE empresa_id=$2 AND id=$3',cuota_tabla) USING a.importe,e,c.id;
 END LOOP;
END $$;

CREATE FUNCTION gf_estado() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE a record; base text:=replace(TG_TABLE_NAME,'_cabecera',''); cuotas text;
BEGIN
 IF TG_OP='INSERT' THEN
  IF NEW.estado<>'BORRADOR' THEN RAISE EXCEPTION 'CREAR_COMO_BORRADOR'; END IF;
  RETURN NEW;
 END IF;
 IF TG_OP='DELETE' THEN
  IF OLD.estado<>'BORRADOR' THEN RAISE EXCEPTION 'COBRO_PAGO_INMUTABLE'; END IF;
  RETURN OLD;
 END IF;
 IF OLD.estado='BORRADOR' AND NEW.estado='CONFIRMADO' THEN
  IF (to_jsonb(NEW)-'estado')<>(to_jsonb(OLD)-'estado') THEN RAISE EXCEPTION 'CONFIRMAR_SIN_EDITAR'; END IF;
  PERFORM gf_aplicar(NEW.empresa_id,base,NEW.id);
 ELSIF OLD.estado='CONFIRMADO' AND NEW.estado='REVERTIDO' AND (to_jsonb(NEW)-'estado')=(to_jsonb(OLD)-'estado') THEN
  PERFORM pg_advisory_xact_lock(NEW.empresa_id);
  IF OLD.apertura_id IS NOT NULL AND EXISTS(SELECT 1 FROM gf_caja_apertura WHERE id=OLD.apertura_id AND cerrada_en IS NOT NULL) THEN RAISE EXCEPTION 'CAJA_CERRADA'; END IF;
  cuotas:=CASE WHEN base='gf_recibo' THEN 'gv_cuota' ELSE 'gc_cuota' END;
  FOR a IN EXECUTE format('SELECT * FROM %I WHERE empresa_id=$1 AND cabecera_id=$2 ORDER BY cuota_id',base||'_aplicacion') USING NEW.empresa_id,NEW.id LOOP
   EXECUTE format('UPDATE %I SET aplicado=aplicado-$1 WHERE empresa_id=$2 AND id=$3',cuotas) USING a.importe,NEW.empresa_id,a.cuota_id;
  END LOOP;
  UPDATE gs_talonario_numero SET estado='ANULADO' WHERE empresa_id=NEW.empresa_id AND id=NEW.numero_id;
 ELSIF OLD.estado<>'BORRADOR' OR NEW.estado<>'BORRADOR' THEN RAISE EXCEPTION 'COBRO_PAGO_INMUTABLE';
 END IF;
 RETURN NEW;
END $$;
CREATE FUNCTION gf_linea_guard() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE h bigint; e bigint; estado text; base text;
BEGIN
 h:=CASE WHEN TG_OP='DELETE' THEN OLD.cabecera_id ELSE NEW.cabecera_id END;
 e:=CASE WHEN TG_OP='DELETE' THEN OLD.empresa_id ELSE NEW.empresa_id END;
 base:=CASE WHEN TG_TABLE_NAME LIKE 'gf_recibo%' THEN 'gf_recibo_cabecera' ELSE 'gf_pago_cabecera' END;
 IF TG_OP='UPDATE' AND (NEW.cabecera_id<>OLD.cabecera_id OR NEW.empresa_id<>OLD.empresa_id) THEN RAISE EXCEPTION 'LINEA_NO_TRANSFERIBLE'; END IF;
 EXECUTE format('SELECT estado FROM %I WHERE empresa_id=$1 AND id=$2 FOR UPDATE',base) INTO estado USING e,h;
 IF estado IS DISTINCT FROM 'BORRADOR' THEN RAISE EXCEPTION 'COBRO_PAGO_INMUTABLE'; END IF;
 IF TG_OP='DELETE' THEN RETURN OLD; END IF;
 RETURN NEW;
END $$;
DO $$ DECLARE b text; BEGIN
 FOREACH b IN ARRAY ARRAY['gf_recibo','gf_pago'] LOOP
  EXECUTE format('CREATE TRIGGER estado BEFORE INSERT OR UPDATE OR DELETE ON %I FOR EACH ROW EXECUTE FUNCTION gf_estado()',b||'_cabecera');
  EXECUTE format('CREATE TRIGGER numero BEFORE INSERT OR UPDATE ON %I FOR EACH ROW EXECUTE FUNCTION gs_numero_exclusivo()',b||'_cabecera');
  EXECUTE format('CREATE TRIGGER detalle BEFORE INSERT OR UPDATE OR DELETE ON %I FOR EACH ROW EXECUTE FUNCTION gf_linea_guard()',b||'_forma');
  EXECUTE format('CREATE TRIGGER detalle BEFORE INSERT OR UPDATE OR DELETE ON %I FOR EACH ROW EXECUTE FUNCTION gf_linea_guard()',b||'_aplicacion');
 END LOOP;
END $$;
CREATE FUNCTION gf_confirmar(e bigint,clase text,documento bigint) RETURNS void LANGUAGE plpgsql AS $$
DECLARE estado text;
BEGIN
 IF clase NOT IN ('gf_recibo','gf_pago') THEN RAISE EXCEPTION 'CLASE_INVALIDA'; END IF;
 PERFORM pg_advisory_xact_lock(e);
 EXECUTE format('SELECT estado FROM %I WHERE empresa_id=$1 AND id=$2 FOR UPDATE',clase||'_cabecera') INTO STRICT estado USING e,documento;
 IF estado='CONFIRMADO' THEN RETURN; END IF;
 EXECUTE format('UPDATE %I SET estado=''CONFIRMADO'' WHERE empresa_id=$1 AND id=$2',clase||'_cabecera') USING e,documento;
END $$;

CREATE FUNCTION gf_cerrar_caja(e bigint,apertura bigint,conteos jsonb) RETURNS void LANGUAGE plpgsql AS $$
DECLARE h gf_caja_apertura; m record; entrada numeric; salida numeric; contado numeric;
BEGIN
 PERFORM pg_advisory_xact_lock(e);
 SELECT * INTO STRICT h FROM gf_caja_apertura WHERE empresa_id=e AND id=apertura FOR UPDATE;
 IF h.cerrada_en IS NOT NULL THEN RAISE EXCEPTION 'CAJA_YA_CERRADA'; END IF;
 IF EXISTS(SELECT 1 FROM gf_recibo_cabecera WHERE empresa_id=e AND apertura_id=apertura AND estado='BORRADOR') OR
 EXISTS(SELECT 1 FROM gf_pago_cabecera WHERE empresa_id=e AND apertura_id=apertura AND estado='BORRADOR') THEN RAISE EXCEPTION 'CAJA_DOCUMENTOS_PENDIENTES'; END IF;
 FOR m IN SELECT * FROM gf_medio_pago WHERE empresa_id=e LOOP
  SELECT coalesce(sum(f.importe),0) INTO entrada FROM gf_recibo_forma f JOIN gf_recibo_cabecera c ON c.id=f.cabecera_id WHERE c.empresa_id=e AND c.apertura_id=apertura AND c.estado='CONFIRMADO' AND f.medio_id=m.id;
  SELECT coalesce(sum(f.importe),0) INTO salida FROM gf_pago_forma f JOIN gf_pago_cabecera c ON c.id=f.cabecera_id WHERE c.empresa_id=e AND c.apertura_id=apertura AND c.estado='CONFIRMADO' AND f.medio_id=m.id;
  IF m.efectivo THEN entrada:=entrada+h.fondo; END IF;
  contado:=(conteos->>m.id::text)::numeric;
  IF contado IS NULL OR contado<0 THEN RAISE EXCEPTION 'FALTA_CONTEO_MEDIO'; END IF;
  INSERT INTO gf_caja_arqueo(empresa_id,apertura_id,medio_id,contado,esperado) VALUES(e,apertura,m.id,contado,entrada-salida);
 END LOOP;
 UPDATE gf_caja_apertura SET cerrada_en=clock_timestamp() WHERE id=apertura;
END $$;
CREATE UNIQUE INDEX un_medio_efectivo ON gf_medio_pago(empresa_id) WHERE efectivo;

CREATE FUNCTION gf_programar_cuotas(e bigint,pref text,documento bigint,plan jsonb) RETURNS void LANGUAGE plpgsql AS $$
DECLARE h record; item jsonb; suma numeric; n bigint; i integer:=0;
BEGIN
 IF pref NOT IN ('gv','gc') THEN RAISE EXCEPTION 'MODULO_INVALIDO'; END IF;
 PERFORM pg_advisory_xact_lock(e);
 EXECUTE format('SELECT * FROM %I WHERE empresa_id=$1 AND id=$2 FOR UPDATE',pref||'_comprobante_cabecera') INTO STRICT h USING e,documento;
 IF h.estado<>'CONFIRMADO' THEN RAISE EXCEPTION 'COMPROBANTE_NO_CONFIRMADO'; END IF;
 EXECUTE format('SELECT count(*) FROM %I WHERE empresa_id=$1 AND comprobante_id=$2 AND (aplicado>0 OR creditado>0)',pref||'_cuota') INTO n USING e,documento;
 IF n>0 THEN RAISE EXCEPTION 'CUOTAS_CON_APLICACIONES'; END IF;
 SELECT sum((value->>'importe')::numeric) INTO suma FROM jsonb_array_elements(plan);
 IF suma IS DISTINCT FROM h.total THEN RAISE EXCEPTION 'CUOTAS_NO_CUADRAN'; END IF;
 EXECUTE format('DELETE FROM %I WHERE empresa_id=$1 AND comprobante_id=$2',pref||'_cuota') USING e,documento;
 FOR item IN SELECT value FROM jsonb_array_elements(plan) LOOP
  i:=i+1;
  EXECUTE format('INSERT INTO %I(empresa_id,comprobante_id,numero,vencimiento,importe) VALUES($1,$2,$3,$4,$5)',pref||'_cuota') USING e,documento,i,(item->>'vencimiento')::date,(item->>'importe')::numeric;
 END LOOP;
END $$;

CREATE FUNCTION gf_usar_credito(e bigint,pref text,credito bigint,cuota bigint,monto numeric,clave uuid) RETURNS bigint LANGUAGE plpgsql AS $$
DECLARE x record; c record; origen record; destino record; usado record; result bigint;
BEGIN
 IF pref NOT IN ('gv','gc') THEN RAISE EXCEPTION 'MODULO_INVALIDO'; END IF;
 PERFORM pg_advisory_xact_lock(e);
 EXECUTE format('SELECT * FROM %I WHERE uid=$1',pref||'_credito_uso') INTO usado USING clave;
 IF usado.id IS NOT NULL THEN
  IF usado.empresa_id<>e OR usado.credito_id<>credito OR usado.cuota_id<>cuota OR usado.importe<>monto THEN RAISE EXCEPTION 'CLAVE_REUTILIZADA'; END IF;
  RETURN usado.id;
 END IF;
 EXECUTE format('SELECT * FROM %I WHERE empresa_id=$1 AND id=$2 FOR UPDATE',pref||'_credito_disponible') INTO STRICT x USING e,credito;
 EXECUTE format('SELECT * FROM %I WHERE empresa_id=$1 AND id=$2 FOR UPDATE',pref||'_cuota') INTO STRICT c USING e,cuota;
 EXECUTE format('SELECT * FROM %I WHERE empresa_id=$1 AND id=$2',pref||'_nota_credito_cabecera') INTO STRICT origen USING e,x.nota_id;
 EXECUTE format('SELECT * FROM %I WHERE empresa_id=$1 AND id=$2',pref||'_comprobante_cabecera') INTO STRICT destino USING e,c.comprobante_id;
 IF origen.estado<>'CONFIRMADO' OR destino.estado<>'CONFIRMADO' OR origen.tercero_id<>destino.tercero_id OR origen.moneda_id<>destino.moneda_id THEN RAISE EXCEPTION 'CREDITO_INCOMPATIBLE'; END IF;
 IF monto<=0 OR monto>x.saldo OR monto>c.saldo THEN RAISE EXCEPTION 'CREDITO_EXCEDE_SALDO'; END IF;
 EXECUTE format('INSERT INTO %I(empresa_id,uid,credito_id,cuota_id,importe) VALUES($1,$2,$3,$4,$5) RETURNING id',pref||'_credito_uso') INTO result USING e,clave,credito,cuota,monto;
 EXECUTE format('UPDATE %I SET utilizado=utilizado+$1 WHERE id=$2',pref||'_credito_disponible') USING monto,credito;
 EXECUTE format('UPDATE %I SET creditado=creditado+$1 WHERE id=$2',pref||'_cuota') USING monto,cuota;
 RETURN result;
END $$;

CREATE FUNCTION gv_devolver(e bigint,linea bigint,cant numeric,clave uuid,motivo_texto text) RETURNS bigint LANGUAGE plpgsql AS $$
DECLARE r gv_devolucion; d gv_comprobante_detalle; m gi_movimiento; result bigint; usado numeric;
BEGIN
 PERFORM pg_advisory_xact_lock(e);
 SELECT * INTO r FROM gv_devolucion WHERE uid=clave;
 IF FOUND THEN
  IF r.empresa_id<>e OR r.detalle_origen_id<>linea OR r.cantidad_base<>cant OR r.motivo<>motivo_texto THEN RAISE EXCEPTION 'CLAVE_REUTILIZADA'; END IF;
  RETURN r.id;
 END IF;
 SELECT * INTO STRICT d FROM gv_comprobante_detalle WHERE empresa_id=e AND id=linea;
 IF NOT EXISTS(SELECT 1 FROM gv_comprobante_cabecera WHERE id=d.cabecera_id AND estado='CONFIRMADO') THEN RAISE EXCEPTION 'ORIGEN_NO_CONFIRMADO'; END IF;
 SELECT * INTO STRICT m FROM gi_movimiento WHERE empresa_id=e AND venta_detalle_id=linea AND motivo='VENTA';
 SELECT coalesce(sum(cantidad_base),0) INTO usado FROM gv_devolucion WHERE empresa_id=e AND detalle_origen_id=linea;
 IF cant<=0 OR usado+cant>d.cantidad_base THEN RAISE EXCEPTION 'DEVOLUCION_EXCESIVA'; END IF;
 IF m.valoracion_pendiente OR m.costo_unitario IS NULL THEN RAISE EXCEPTION 'COSTO_ORIGINAL_PENDIENTE'; END IF;
 result:=gi_postear(e,gen_random_uuid(),d.posicion_id,cant,m.costo_unitario,false,'VENTA',linea,now(),'DEVOLUCION');
 INSERT INTO gv_devolucion(empresa_id,uid,detalle_origen_id,cantidad_base,movimiento_id,motivo) VALUES(e,clave,linea,cant,result,motivo_texto) RETURNING id INTO result;
 RETURN result;
END $$;

CREATE FUNCTION gc_devolver(e bigint,linea bigint,cant numeric,clave uuid,motivo_texto text) RETURNS bigint LANGUAGE plpgsql AS $$
DECLARE r gc_devolucion; d gc_recepcion_detalle; m gi_movimiento; result bigint; usado numeric;
BEGIN
 PERFORM pg_advisory_xact_lock(e);
 SELECT * INTO r FROM gc_devolucion WHERE uid=clave;
 IF FOUND THEN
  IF r.empresa_id<>e OR r.detalle_origen_id<>linea OR r.cantidad_base<>cant OR r.motivo<>motivo_texto THEN RAISE EXCEPTION 'CLAVE_REUTILIZADA'; END IF;
  RETURN r.id;
 END IF;
 SELECT * INTO STRICT d FROM gc_recepcion_detalle WHERE empresa_id=e AND id=linea;
 IF NOT EXISTS(SELECT 1 FROM gc_recepcion_cabecera WHERE id=d.cabecera_id AND estado='CONFIRMADO') THEN RAISE EXCEPTION 'ORIGEN_NO_CONFIRMADO'; END IF;
 SELECT * INTO STRICT m FROM gi_movimiento WHERE empresa_id=e AND recepcion_detalle_id=linea AND motivo='RECEPCION';
 SELECT coalesce(sum(cantidad_base),0) INTO usado FROM gc_devolucion WHERE empresa_id=e AND detalle_origen_id=linea;
 IF cant<=0 OR usado+cant>d.cantidad_base THEN RAISE EXCEPTION 'DEVOLUCION_EXCESIVA'; END IF;
 IF m.valoracion_pendiente OR m.costo_unitario IS NULL THEN RAISE EXCEPTION 'COSTO_ORIGINAL_PENDIENTE'; END IF;
 result:=gi_postear(e,gen_random_uuid(),d.posicion_id,-cant,m.costo_unitario,false,'RECEPCION',linea,now(),'DEVOLUCION_PROVEEDOR');
 INSERT INTO gc_devolucion(empresa_id,uid,detalle_origen_id,cantidad_base,movimiento_id,motivo) VALUES(e,clave,linea,cant,result,motivo_texto) RETURNING id INTO result;
 RETURN result;
END $$;

CREATE FUNCTION gs_anular(e bigint,modulo text,documento bigint,u bigint,motivo_texto text) RETURNS void LANGUAGE plpgsql AS $$
DECLARE h record; m record; n bigint; campo text; movimientos text; pref text; cr record;
BEGIN
 IF modulo NOT IN ('gc_orden_compra','gc_recepcion','gc_comprobante','gv_cotizacion','gv_pedido','gv_comprobante','gi_ajuste','gi_transferencia','gl_despacho','gv_nota_credito','gc_nota_credito') THEN RAISE EXCEPTION 'MODULO_INVALIDO'; END IF;
 IF length(trim(motivo_texto))=0 THEN RAISE EXCEPTION 'MOTIVO_REQUERIDO'; END IF;
 PERFORM pg_advisory_xact_lock(e);
 EXECUTE format('SELECT * FROM %I WHERE empresa_id=$1 AND id=$2 FOR UPDATE',modulo||'_cabecera') INTO STRICT h USING e,documento;
 PERFORM gs_acceso(e,u,h.sucursal_id);
 IF h.estado='REVERTIDO' THEN RETURN; END IF;
 IF h.estado<>'CONFIRMADO' THEN RAISE EXCEPTION 'NO_CONFIRMADO'; END IF;
 IF modulo IN ('gv_nota_credito','gc_nota_credito') THEN
  pref:=left(modulo,2);
  EXECUTE format('SELECT count(*) FROM %I WHERE empresa_id=$1 AND nota_id=$2 AND utilizado>0',pref||'_credito_disponible') INTO n USING e,documento;
  IF n>0 THEN RAISE EXCEPTION 'CREDITO_YA_UTILIZADO'; END IF;
  FOR cr IN EXECUTE format('SELECT * FROM %I WHERE empresa_id=$1 AND nota_id=$2 ORDER BY cuota_id',pref||'_credito_aplicacion') USING e,documento LOOP
   EXECUTE format('UPDATE %I SET creditado=creditado-$1 WHERE empresa_id=$2 AND id=$3',pref||'_cuota') USING cr.importe,e,cr.cuota_id;
  END LOOP;
 END IF;
 IF modulo IN ('gv_comprobante','gc_comprobante') THEN
  EXECUTE format('SELECT count(*) FROM %I WHERE empresa_id=$1 AND comprobante_id=$2 AND (aplicado>0 OR creditado>0)',left(modulo,2)||'_cuota') INTO n USING e,documento;
  IF n>0 THEN RAISE EXCEPTION 'DESAPLICAR_COBROS_PAGOS'; END IF;
  EXECUTE format('SELECT count(*) FROM %I WHERE empresa_id=$1 AND comprobante_origen_id=$2 AND estado=''CONFIRMADO''',left(modulo,2)||'_nota_credito_cabecera') INTO n USING e,documento;
  IF n>0 THEN RAISE EXCEPTION 'RESOLVER_NOTAS_CREDITO'; END IF;
 END IF;
 IF modulo='gv_comprobante' AND EXISTS(SELECT 1 FROM gl_asignacion_venta x JOIN gv_comprobante_detalle d ON d.id=x.venta_detalle_id WHERE d.cabecera_id=documento AND d.empresa_id=e) THEN RAISE EXCEPTION 'RESOLVER_DESPACHOS'; END IF;
 IF modulo='gc_recepcion' AND EXISTS(SELECT 1 FROM gc_factura_recepcion_detalle x JOIN gc_recepcion_detalle d ON d.id=x.recepcion_detalle_id JOIN gc_comprobante_detalle f ON f.id=x.factura_detalle_id JOIN gc_comprobante_cabecera c ON c.id=f.cabecera_id WHERE d.cabecera_id=documento AND d.empresa_id=e AND c.estado<>'REVERTIDO') THEN RAISE EXCEPTION 'RESOLVER_FACTURAS'; END IF;
 IF modulo='gv_comprobante' AND EXISTS(SELECT 1 FROM gv_devolucion x JOIN gv_comprobante_detalle d ON d.id=x.detalle_origen_id WHERE d.empresa_id=e AND d.cabecera_id=documento) THEN RAISE EXCEPTION 'RESOLVER_DEVOLUCIONES'; END IF;
 IF modulo='gc_recepcion' AND EXISTS(SELECT 1 FROM gc_devolucion x JOIN gc_recepcion_detalle d ON d.id=x.detalle_origen_id WHERE d.empresa_id=e AND d.cabecera_id=documento) THEN RAISE EXCEPTION 'RESOLVER_DEVOLUCIONES'; END IF;
 campo:=CASE modulo WHEN 'gv_comprobante' THEN 'venta_detalle_id' WHEN 'gc_recepcion' THEN 'recepcion_detalle_id' WHEN 'gi_ajuste' THEN 'ajuste_detalle_id' WHEN 'gi_transferencia' THEN 'transferencia_detalle_id' END;
 IF campo IS NOT NULL THEN
  movimientos:=format('SELECT m.* FROM gi_movimiento m JOIN %I d ON d.id=m.%I WHERE d.empresa_id=$1 AND d.cabecera_id=$2 ORDER BY m.id DESC',modulo||'_detalle',campo);
 ELSIF modulo='gc_comprobante' THEN
  movimientos:='SELECT m.* FROM gi_movimiento m JOIN gc_factura_recepcion_detalle x ON x.id=m.factura_recepcion_id JOIN gc_comprobante_detalle d ON d.id=x.factura_detalle_id WHERE d.empresa_id=$1 AND d.cabecera_id=$2 ORDER BY m.id DESC';
 ELSE movimientos:='SELECT * FROM gi_movimiento WHERE empresa_id=$1 AND id=$2 AND false';
 END IF;
 FOR m IN EXECUTE movimientos USING e,documento LOOP
  IF m.valoracion_pendiente THEN RAISE EXCEPTION 'RESOLVER_COSTO_PENDIENTE'; END IF;
  PERFORM gi_postear(e,gen_random_uuid(),m.posicion_id,-m.cantidad,m.costo_unitario,false,'REVERSION',m.id,now(),'ANULACION',-m.valor_inventario,-m.diferencia_consumida);
 END LOOP;
 EXECUTE format('INSERT INTO %I(empresa_id,cabecera_id,usuario_id,motivo) VALUES($1,$2,$3,$4)',modulo||'_anulacion') USING e,documento,u,motivo_texto;
 EXECUTE format('UPDATE %I SET estado=''REVERTIDO'' WHERE empresa_id=$1 AND id=$2',modulo||'_cabecera') USING e,documento;
 IF h.numero_id IS NOT NULL THEN UPDATE gs_talonario_numero SET estado='ANULADO' WHERE empresa_id=e AND id=h.numero_id; END IF;
END $$;

CREATE FUNCTION gc_compra_directa(e bigint,factura bigint,recepcion bigint) RETURNS void LANGUAGE plpgsql AS $$
DECLARE f gc_comprobante_cabecera; r gc_recepcion_cabecera; l record; n bigint;
BEGIN
 PERFORM pg_advisory_xact_lock(e);
 SELECT * INTO STRICT f FROM gc_comprobante_cabecera WHERE empresa_id=e AND id=factura FOR UPDATE;
 SELECT * INTO STRICT r FROM gc_recepcion_cabecera WHERE empresa_id=e AND id=recepcion FOR UPDATE;
 IF f.tercero_id<>r.tercero_id OR f.sucursal_id<>r.sucursal_id THEN RAISE EXCEPTION 'COMPRA_DIRECTA_INCOMPATIBLE'; END IF;
 IF f.estado='CONFIRMADO' AND r.estado='CONFIRMADO' THEN
  IF NOT EXISTS(SELECT 1 FROM gc_factura_recepcion_detalle x JOIN gc_comprobante_detalle a ON a.id=x.factura_detalle_id JOIN gc_recepcion_detalle b ON b.id=x.recepcion_detalle_id WHERE a.cabecera_id=factura AND b.cabecera_id=recepcion) THEN RAISE EXCEPTION 'DOCUMENTOS_NO_VINCULADOS'; END IF;
  RETURN;
 END IF;
 IF f.estado<>'BORRADOR' OR r.estado<>'BORRADOR' THEN RAISE EXCEPTION 'COMPRA_DIRECTA_REQUIERE_BORRADORES'; END IF;
 SELECT count(*) INTO n FROM gc_recepcion_detalle d WHERE d.cabecera_id=recepcion AND NOT EXISTS(SELECT 1 FROM gc_comprobante_detalle a WHERE a.cabecera_id=factura AND a.renglon=d.renglon AND a.articulo_id=d.articulo_id AND a.cantidad_base=d.cantidad_base);
 IF n>0 THEN RAISE EXCEPTION 'LINEAS_NO_COINCIDEN'; END IF;
 PERFORM gs_confirmar(e,'gc_recepcion',recepcion);
 FOR l IN SELECT a.*,b.id rid FROM gc_comprobante_detalle a JOIN gi_articulo art ON art.id=a.articulo_id LEFT JOIN gc_recepcion_detalle b ON b.cabecera_id=recepcion AND b.renglon=a.renglon AND b.articulo_id=a.articulo_id AND b.cantidad_base=a.cantidad_base WHERE a.cabecera_id=factura AND art.mueve_stock LOOP
  IF l.rid IS NULL OR l.costo_unitario_base IS NULL THEN RAISE EXCEPTION 'LINEA_SIN_RECEPCION_O_COSTO'; END IF;
  INSERT INTO gc_factura_recepcion_detalle(empresa_id,factura_detalle_id,recepcion_detalle_id,cantidad_base,costo_final_base) VALUES(e,l.id,l.rid,l.cantidad_base,l.costo_unitario_base);
 END LOOP;
 PERFORM gs_confirmar(e,'gc_comprobante',factura);
END $$;

CREATE FUNCTION gi_reservar(e bigint,pos bigint,cant numeric,pedido bigint DEFAULT NULL,terminal bigint DEFAULT NULL) RETURNS bigint LANGUAGE plpgsql AS $$
DECLARE s gi_existencia; r bigint;
BEGIN
 PERFORM pg_advisory_xact_lock(e);
 INSERT INTO gi_existencia(empresa_id,posicion_id) VALUES(e,pos) ON CONFLICT DO NOTHING;
 SELECT * INTO STRICT s FROM gi_existencia WHERE empresa_id=e AND posicion_id=pos FOR UPDATE;
 IF cant<=0 OR s.disponible<cant THEN RAISE EXCEPTION 'RESERVA_SIN_DISPONIBILIDAD'; END IF;
 INSERT INTO gi_reserva(empresa_id,posicion_id,pedido_detalle_id,terminal_id,cantidad) VALUES(e,pos,pedido,terminal,cant) RETURNING id INTO r;
 UPDATE gi_existencia SET reservada=reservada+cant WHERE id=s.id;
 RETURN r;
END $$;
CREATE FUNCTION gi_liberar_reserva(e bigint,r bigint) RETURNS void LANGUAGE plpgsql AS $$
DECLARE x gi_reserva;
BEGIN
 PERFORM pg_advisory_xact_lock(e);
 SELECT * INTO STRICT x FROM gi_reserva WHERE empresa_id=e AND id=r FOR UPDATE;
 IF NOT x.activa THEN RETURN; END IF;
 UPDATE gi_existencia SET reservada=reservada-(x.cantidad-x.consumida) WHERE empresa_id=e AND posicion_id=x.posicion_id;
 UPDATE gi_reserva SET activa=false WHERE id=r;
END $$;

CREATE FUNCTION gi_resolver_valoracion(e bigint,linea_ajuste bigint,u bigint,valor_objetivo numeric,diferencia_consumida numeric,motivo_texto text) RETURNS bigint LANGUAGE plpgsql AS $$
DECLARE d gi_ajuste_detalle; h gi_ajuste_cabecera; v gi_valoracion; result bigint;
BEGIN
 PERFORM pg_advisory_xact_lock(e);
 SELECT * INTO STRICT d FROM gi_ajuste_detalle WHERE empresa_id=e AND id=linea_ajuste;
 SELECT * INTO STRICT h FROM gi_ajuste_cabecera WHERE empresa_id=e AND id=d.cabecera_id;
 PERFORM gs_acceso(e,u,h.sucursal_id);
 IF NOT EXISTS(SELECT 1 FROM gs_usuario_sucursal WHERE empresa_id=e AND sucursal_id=h.sucursal_id AND usuario_id=u AND supervisor) THEN RAISE EXCEPTION 'REQUIERE_SUPERVISOR'; END IF;
 IF h.estado<>'CONFIRMADO' OR h.efecto_stock_aplicado<>'NINGUNO' OR d.posicion_id IS NULL OR length(trim(motivo_texto))=0 THEN RAISE EXCEPTION 'AJUSTE_VALOR_INVALIDO'; END IF;
 SELECT * INTO STRICT v FROM gi_valoracion WHERE empresa_id=e AND articulo_id=d.articulo_id FOR UPDATE;
 IF NOT v.pendiente OR EXISTS(SELECT 1 FROM gi_movimiento WHERE ajuste_detalle_id=d.id AND motivo='RESOLUCION_COSTO') THEN RAISE EXCEPTION 'SIN_PENDIENTE_O_YA_RESUELTO'; END IF;
 IF (v.cantidad=0 AND valor_objetivo<>0) OR (v.cantidad<>0 AND valor_objetivo/v.cantidad<0) THEN RAISE EXCEPTION 'VALOR_OBJETIVO_INVALIDO'; END IF;
 result:=gi_postear(e,gen_random_uuid(),d.posicion_id,0,NULL,false,'AJUSTE',d.id,now(),'RESOLUCION_COSTO',valor_objetivo-v.valor,diferencia_consumida);
 UPDATE gi_valoracion SET pendiente=false,ultimo_costo=CASE WHEN cantidad<>0 THEN valor_objetivo/cantidad ELSE ultimo_costo END WHERE id=v.id;
 INSERT INTO gs_auditoria(empresa_id,usuario_id,tabla,registro_id,accion,antes,despues) VALUES(e,u,'gi_valoracion',v.id,'RESOLVER_COSTO',to_jsonb(v),jsonb_build_object('movimiento_id',result,'valor_objetivo',valor_objetivo,'diferencia_consumida',diferencia_consumida,'motivo',motivo_texto));
 RETURN result;
END $$;

CREATE FUNCTION gp_validar_extension() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE p gp_permiso_offline; s gp_permiso_offline;
BEGIN
 SELECT * INTO STRICT p FROM gp_permiso_offline WHERE empresa_id=NEW.empresa_id AND id=NEW.permiso_id;
 SELECT * INTO STRICT s FROM gp_permiso_offline WHERE empresa_id=NEW.empresa_id AND id=NEW.supervisor_permiso_id;
 IF NOT s.supervisor OR p.terminal_id<>s.terminal_id OR NEW.hasta<=p.vence_en OR NEW.hasta>p.vence_en+interval '24 hours'
 OR s.vence_en<p.vence_en OR EXISTS(SELECT 1 FROM gp_extension_permiso WHERE empresa_id=NEW.empresa_id AND permiso_id=p.id)
 THEN RAISE EXCEPTION 'EXTENSION_INVALIDA'; END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER extension_validar BEFORE INSERT ON gp_extension_permiso FOR EACH ROW EXECUTE FUNCTION gp_validar_extension();
CREATE TRIGGER extension_inmutable BEFORE UPDATE OR DELETE ON gp_extension_permiso FOR EACH ROW EXECUTE FUNCTION gs_inmutable();

CREATE FUNCTION gp_integrar_venta(e bigint,terminal bigint,secuencia bigint,clave uuid,ocurrido timestamptz,datos jsonb)
RETURNS bigint LANGUAGE plpgsql AS $$
DECLARE ev gp_evento_entrada; h bigint; num bigint; lin jsonb; permiso gp_permiso_offline; vence timestamptz; suc bigint; err text;
BEGIN
 PERFORM pg_advisory_xact_lock(e);
 SELECT * INTO ev FROM gp_evento_entrada WHERE uid=clave;
 IF FOUND THEN
  IF ev.empresa_id<>e OR ev.terminal_id<>terminal OR ev.secuencia<>secuencia OR ev.ocurrido_en<>ocurrido OR ev.contenido<>datos THEN RAISE EXCEPTION 'CLAVE_REUTILIZADA'; END IF;
  IF ev.estado='APLICADO' THEN RETURN ev.venta_id; END IF;
 ELSE
  INSERT INTO gp_evento_entrada(empresa_id,uid,terminal_id,secuencia,ocurrido_en,contenido) VALUES(e,clave,terminal,secuencia,ocurrido,datos) RETURNING * INTO ev;
 END IF;
 BEGIN
  SELECT sucursal_id INTO STRICT suc FROM gp_terminal WHERE empresa_id=e AND id=terminal AND activa;
  SELECT * INTO STRICT permiso FROM gp_permiso_offline WHERE empresa_id=e AND id=(datos->>'permiso_id')::bigint AND terminal_id=terminal AND usuario_id=(datos->>'usuario_id')::bigint;
  SELECT greatest(permiso.vence_en,coalesce(max(hasta),permiso.vence_en)) INTO vence FROM gp_extension_permiso WHERE empresa_id=e AND permiso_id=permiso.id;
  IF ocurrido<permiso.emitido_en OR ocurrido>vence OR ocurrido>now()+interval '5 minutes' THEN RAISE EXCEPTION 'PERMISO_OFFLINE_VENCIDO'; END IF;
  num:=gs_asignar_numero(e,(datos->>'talonario_id')::bigint,permiso.usuario_id,clave,terminal,(datos->>'rango_id')::bigint,(datos->>'numero')::bigint,ocurrido);
  INSERT INTO gv_comprobante_cabecera(empresa_id,uid,sucursal_id,usuario_id,tipo_id,numero_id,terminal_id,tercero_id,moneda_id,cambio_base,fecha_operacion,origen,costo_local)
  VALUES(e,clave,suc,permiso.usuario_id,(datos->>'tipo_id')::bigint,num,terminal,(datos->>'cliente_id')::bigint,(datos->>'moneda_id')::bigint,
  coalesce((datos->>'cambio_base')::numeric,1),ocurrido,'POS',(datos->>'costo_local')::numeric) RETURNING id INTO h;
  FOR lin IN SELECT value FROM jsonb_array_elements(datos->'lineas') LOOP
   INSERT INTO gv_comprobante_detalle(empresa_id,cabecera_id,renglon,articulo_id,presentacion_id,descripcion,cantidad,factor_base,precio,descuento,tasa_impuesto,base_imponible,impuesto,posicion_id)
   VALUES(e,h,(lin->>'renglon')::int,(lin->>'articulo_id')::bigint,(lin->>'presentacion_id')::bigint,lin->>'descripcion',(lin->>'cantidad')::numeric,
   (lin->>'factor_base')::numeric,(lin->>'precio')::numeric,coalesce((lin->>'descuento')::numeric,0),coalesce((lin->>'tasa_impuesto')::numeric,0),coalesce((lin->>'base_imponible')::numeric,0),coalesce((lin->>'impuesto')::numeric,0),(lin->>'posicion_id')::bigint);
  END LOOP;
  PERFORM gs_confirmar(e,'gv_comprobante',h);
  UPDATE gp_evento_entrada SET estado='APLICADO',venta_id=h,error=NULL WHERE id=ev.id;
  RETURN h;
 EXCEPTION WHEN OTHERS THEN
  GET STACKED DIAGNOSTICS err=MESSAGE_TEXT;
  UPDATE gp_evento_entrada SET estado='CONFLICTO',error=err WHERE id=ev.id;
  RETURN NULL;
 END;
END $$;

-- Indices operativos y vistas de conciliacion.
CREATE INDEX kardex_articulo_fecha ON gi_movimiento(empresa_id,articulo_id,registrado_en,id);
CREATE UNIQUE INDEX salida_venta_unica ON gi_movimiento(empresa_id,venta_detalle_id) WHERE motivo='VENTA';
CREATE UNIQUE INDEX entrada_recepcion_unica ON gi_movimiento(empresa_id,recepcion_detalle_id) WHERE motivo='RECEPCION';
CREATE INDEX cola_pendiente ON gp_evento_entrada(empresa_id,terminal_id,secuencia) WHERE estado<>'APLICADO';
CREATE VIEW gi_conciliacion AS
 SELECT v.empresa_id,v.articulo_id,v.cantidad,v.valor,
 coalesce((SELECT sum(x.cantidad) FROM gi_existencia x JOIN gi_posicion p ON p.id=x.posicion_id WHERE p.empresa_id=v.empresa_id AND p.articulo_id=v.articulo_id),0) cantidad_posiciones,
 coalesce((SELECT sum(m.cantidad) FROM gi_movimiento m WHERE m.empresa_id=v.empresa_id AND m.articulo_id=v.articulo_id),0) cantidad_kardex,
 coalesce((SELECT sum(m.valor_inventario) FROM gi_movimiento m WHERE m.empresa_id=v.empresa_id AND m.articulo_id=v.articulo_id),0) valor_kardex
 FROM gi_valoracion v;
DO $$ DECLARE r record; i integer:=0; BEGIN
 FOR r IN SELECT c.conrelid::regclass tabla, string_agg(quote_ident(a.attname),',' ORDER BY k.ord) cols
 FROM pg_constraint c CROSS JOIN LATERAL unnest(c.conkey) WITH ORDINALITY k(attnum,ord)
 JOIN pg_attribute a ON a.attrelid=c.conrelid AND a.attnum=k.attnum
 WHERE c.contype='f' AND c.connamespace='erp_v4'::regnamespace GROUP BY c.oid,c.conrelid LOOP
  i:=i+1; EXECUTE format('CREATE INDEX %I ON %s(%s)','v4_fk_'||i,r.tabla,r.cols);
 END LOOP;
END $$;
REVOKE ALL ON ALL TABLES IN SCHEMA erp_v4 FROM PUBLIC;
DO $$ DECLARE r record; BEGIN
 FOR r IN SELECT oid::regprocedure f FROM pg_proc WHERE pronamespace='erp_v4'::regnamespace LOOP
  EXECUTE format('ALTER FUNCTION %s SET search_path=erp_v4,pg_catalog',r.f);
 END LOOP;
END $$;
REVOKE EXECUTE ON ALL FUNCTIONS IN SCHEMA erp_v4 FROM PUBLIC;

COMMIT;
