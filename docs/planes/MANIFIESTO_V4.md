# Manifiesto V4 inicial

| Objeto | Propietario lógico | Procedencia |
|---|---|---|
| gs_moneda | plataforma/identidad | V4 |
| gs_empresa | plataforma/identidad | V4 |
| gs_usuario | plataforma/identidad | V4 |
| gs_usuario_empresa | plataforma/identidad | V4 |
| gs_sucursal | plataforma/identidad | V4 |
| gs_rol | plataforma/identidad | V4 |
| gs_permiso | plataforma/identidad | V4 |
| gs_rol_permiso | plataforma/identidad | V4 |
| gs_usuario_rol | plataforma/identidad | V4 |
| gs_modulo | plataforma/identidad | V4 |
| gs_empresa_modulo | plataforma/identidad | V4 |
| gs_cotizacion_moneda | plataforma/identidad | V4 |
| gs_entidad | plataforma/identidad | V4 |
| gs_entidad_documento | plataforma/identidad | V4 |
| gs_entidad_contacto | plataforma/identidad | V4 |
| gs_entidad_direccion | plataforma/identidad | V4 |
| gf_condicion_pago | tesoreria | V4 |
| gf_condicion_pago_cuota | tesoreria | V4 |
| gv_cliente | ventas | V4 |
| gc_proveedor | compras | V4 |
| gi_unidad_medida | inventario/catalogo | V4 |
| gi_categoria | inventario/catalogo | V4 |
| gi_producto | inventario/catalogo | V4 |
| gi_articulo | inventario/catalogo | V4 |
| gi_presentacion | inventario/catalogo | V4 |
| gi_codigo_barra | inventario/catalogo | V4 |
| gp_regla_balanza | pos/numeracion | V4 |
| gi_kit_componente | inventario/catalogo | V4 |
| gi_lote | inventario/catalogo | V4 |
| gi_serie | inventario/catalogo | V4 |
| gs_impuesto | plataforma/identidad | V4 |
| gs_impuesto_tasa | plataforma/identidad | V4 |
| gi_articulo_impuesto | inventario/catalogo | V4 |
| gv_lista_precio | ventas | V4 |
| gv_precio | ventas | V4 |
| gv_promocion | ventas | V4 |
| gv_promocion_articulo | ventas | V4 |
| gs_usuario_sucursal | plataforma/identidad | V4 |
| gp_terminal | pos/numeracion | V4 |
| gs_modalidad_emision | plataforma/identidad | V4 |
| gs_habilitacion_fiscal | plataforma/identidad | V4 |
| gs_clase_documento | plataforma/identidad | V4 |
| gs_tipo_comprobante | plataforma/identidad | V4 |
| gs_talonario | plataforma/identidad | V4 |
| gs_talonario_usuario | plataforma/identidad | V4 |
| gs_talonario_alternativo | plataforma/identidad | V4 |
| gs_talonario_rango_terminal | plataforma/identidad | V4 |
| gs_talonario_numero | plataforma/identidad | V4 |
| gi_deposito | inventario/catalogo | V4 |
| gi_ubicacion | inventario/catalogo | V4 |
| gi_posicion | inventario/catalogo | V4 |
| gi_existencia | inventario/catalogo | V4 |
| gi_valoracion | inventario/catalogo | V4 |
| gc_orden_compra_cabecera | compras | V4 |
| gc_orden_compra_detalle | compras | V4 |
| gc_recepcion_cabecera | compras | V4 |
| gc_recepcion_detalle | compras | V4 |
| gc_comprobante_cabecera | compras | V4 |
| gc_comprobante_detalle | compras | V4 |
| gv_cotizacion_cabecera | ventas | V4 |
| gv_cotizacion_detalle | ventas | V4 |
| gv_pedido_cabecera | ventas | V4 |
| gv_pedido_detalle | ventas | V4 |
| gv_comprobante_cabecera | ventas | V4 |
| gv_comprobante_detalle | ventas | V4 |
| gi_ajuste_cabecera | inventario/catalogo | V4 |
| gi_ajuste_detalle | inventario/catalogo | V4 |
| gi_transferencia_cabecera | inventario/catalogo | V4 |
| gi_transferencia_detalle | inventario/catalogo | V4 |
| gl_despacho_cabecera | logistica | V4 |
| gl_despacho_detalle | logistica | V4 |
| gv_nota_credito_cabecera | ventas | V4 |
| gv_nota_credito_detalle | ventas | V4 |
| gc_nota_credito_cabecera | compras | V4 |
| gc_nota_credito_detalle | compras | V4 |
| gc_recepcion_orden | compras | V4 |
| gc_factura_recepcion_detalle | compras | V4 |
| gv_pedido_factura | ventas | V4 |
| gl_asignacion_venta | logistica | V4 |
| gl_entrega_evento | logistica | V4 |
| gi_movimiento | inventario/catalogo | V4 |
| gi_incidencia | inventario/catalogo | V4 |
| gi_reserva | inventario/catalogo | V4 |
| gv_devolucion | ventas | V4 |
| gc_devolucion | compras | V4 |
| gc_orden_compra_anulacion | compras | V4 |
| gc_recepcion_anulacion | compras | V4 |
| gc_comprobante_anulacion | compras | V4 |
| gv_cotizacion_anulacion | ventas | V4 |
| gv_pedido_anulacion | ventas | V4 |
| gv_comprobante_anulacion | ventas | V4 |
| gi_ajuste_anulacion | inventario/catalogo | V4 |
| gi_transferencia_anulacion | inventario/catalogo | V4 |
| gl_despacho_anulacion | logistica | V4 |
| gv_nota_credito_anulacion | ventas | V4 |
| gc_nota_credito_anulacion | compras | V4 |
| gf_caja | tesoreria | V4 |
| gf_caja_apertura | tesoreria | V4 |
| gf_medio_pago | tesoreria | V4 |
| gv_cuota | ventas | V4 |
| gc_cuota | compras | V4 |
| gv_credito_aplicacion | ventas | V4 |
| gv_credito_disponible | ventas | V4 |
| gv_credito_uso | ventas | V4 |
| gc_credito_aplicacion | compras | V4 |
| gc_credito_disponible | compras | V4 |
| gc_credito_uso | compras | V4 |
| gf_recibo_cabecera | tesoreria | V4 |
| gf_recibo_forma | tesoreria | V4 |
| gf_recibo_aplicacion | tesoreria | V4 |
| gf_pago_cabecera | tesoreria | V4 |
| gf_pago_forma | tesoreria | V4 |
| gf_pago_aplicacion | tesoreria | V4 |
| gf_caja_arqueo | tesoreria | V4 |
| gf_cuenta_bancaria | tesoreria | V4 |
| gf_movimiento_bancario | tesoreria | V4 |
| gp_permiso_offline | pos/numeracion | V4 |
| gp_extension_permiso | pos/numeracion | V4 |
| gp_evento_entrada | pos/numeracion | V4 |
| gs_evento_salida | plataforma/identidad | V4 |
| gs_auditoria | plataforma/identidad | V4 |
| fe_documento | revisar | V4 |
| fe_evento | revisar | V4 |
| gs_acceso() | revisar (función V4) | V4 |
| gs_validar_talonario() | revisar (función V4) | V4 |
| gs_reservar_rango() | revisar (función V4) | V4 |
| gs_asignar_numero() | revisar (función V4) | V4 |
| gs_numero_disponible() | revisar (función V4) | V4 |
| gs_rango_guard() | revisar (función V4) | V4 |
| gs_talonario_guard() | revisar (función V4) | V4 |
| gs_numero_exclusivo() | revisar (función V4) | V4 |
| gi_validar_posicion() | revisar (función V4) | V4 |
| gi_postear() | revisar (función V4) | V4 |
| gs_inmutable() | revisar (función V4) | V4 |
| gs_validar_numero_edicion() | revisar (función V4) | V4 |
| gc_aplicar_diferencia() | revisar (función V4) | V4 |
| gc_vinculo_guard() | revisar (función V4) | V4 |
| gc_aplicar_vinculos_posteriores() | revisar (función V4) | V4 |
| gs_vinculo_cantidades() | revisar (función V4) | V4 |
| gs_confirmacion() | revisar (función V4) | V4 |
| gs_linea_borrador() | revisar (función V4) | V4 |
| gs_confirmar() | revisar (función V4) | V4 |
| gf_aplicar() | revisar (función V4) | V4 |
| gf_estado() | revisar (función V4) | V4 |
| gf_linea_guard() | revisar (función V4) | V4 |
| gf_confirmar() | revisar (función V4) | V4 |
| gf_cerrar_caja() | revisar (función V4) | V4 |
| gf_programar_cuotas() | revisar (función V4) | V4 |
| gf_usar_credito() | revisar (función V4) | V4 |
| gv_devolver() | revisar (función V4) | V4 |
| gc_devolver() | revisar (función V4) | V4 |
| gs_anular() | revisar (función V4) | V4 |
| gc_compra_directa() | revisar (función V4) | V4 |
| gi_reservar() | revisar (función V4) | V4 |
| gi_liberar_reserva() | revisar (función V4) | V4 |
| gi_resolver_valoracion() | revisar (función V4) | V4 |
| gp_validar_extension() | revisar (función V4) | V4 |
| gp_integrar_venta() | revisar (función V4) | V4 |

## Regla CI

Los módulos no importan `internal` de otro módulo. Toda tabla o función nueva requiere propietario único y revisión de dependencias.
