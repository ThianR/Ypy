# Operación, actualización y recuperación

Complementa [04_VALIDACION_Y_PILOTO.md](04_VALIDACION_Y_PILOTO.md); sus requisitos de respaldo, capacidad y aceptación siguen vigentes.

## Diagnóstico

Implementar liveness (proceso activo), readiness central (puede recibir trabajo) y estado local (puede operar aislado). Falta de conexión central no significa que el POS local esté averiado. Datos sensibles de estado requieren autorización; no exponer endpoints de soporte a toda la red.

Estado de terminal: versión de aplicación/contrato/esquema, instalación, sesión de caja, espacio disponible, pendientes/conflictos, última sincronización exitosa y rangos restantes. Logs estructurados por módulo, operación, empresa y terminal. Evitar servicios de observabilidad externos obligatorios para iniciar; producir datos utilizables por soporte.

## Tiempo y día comercial

Persistir instantes como timestamptz en PostgreSQL y transportarlos con zona/offset explícito, normalizados a UTC para comparación técnica. Guardar contexto de zona IANA comercial; inicialmente America/Asuncion, con configuración de empresa/sucursal. No fijar un offset numérico eterno.

Distinguir ocurrido_en (dispositivo), recibido_en (servidor), procesado_en (integración) y fecha_comercial. Fecha comercial propuesta: la de apertura del turno para operaciones de esa sesión; una sucursal puede configurar un corte horario, aplicado al abrir la sesión. Conservar la política/version usada. No recalcular cierres históricos al cambiar zona o corte. La fecha fiscal sigue las reglas de su modalidad, no se sustituye libremente por fecha comercial.

El orden de integración depende de secuencia/cursor y transacciones; no de confiar en el reloj de la caja. Medir desvíos y registrar alertas. F02/F04 revisan campos existentes y pruebas con turno que cruza medianoche y reloj atrasado.

## Actualizaciones

Artefactos de versión identificable e integridad/autenticidad verificables. Instalar fuera de una operación activa, con migración local ensayada, respaldo consistente y pendientes preservados. El mecanismo puede comenzar como instalador administrado: no exigir un segundo ejecutable actualizador ni actualizaciones remotas automáticas para el MVP.

No volver a un binario anterior si no entiende el esquema actualizado. Separar reversión de aplicación compatible y recuperación de datos; nunca restaurar un respaldo viejo descartando ventas emitidas para simular rollback. En PWA, coordinar pestañas antes de activar service worker o migración de IndexedDB.

## Recuperación por incidente

| Incidente | Procedimiento requerido |
|---|---|
| Respuesta central perdida | Reenviar mismo UUID; consultar resultado persistido |
| Proceso local cerrado | Recuperar journal; ENVIANDO vuelve a estado reintentable |
| Datos del navegador borrados | Suspender esa identidad, recuperar desde copia/central lo disponible y conciliar números; declarar posible pérdida de pendientes |
| PC reemplazada | Enrolar nueva instalación; inmovilizar anterior; conciliar antes de habilitar nuevos rangos |
| Base central restaurada | Conciliar operaciones locales y acuses posteriores al respaldo antes de purgar/reutilizar registros |
| Impresión incierta | Mantener venta; reimpresión explícita identificada |
| Conflicto comercial | Resolver con permiso/motivo y recibo central; no borrar cola ni editar original |

Definir RPO/RTO y retención de evidencias con el piloto. Respaldos consistentes de SQLite, exportación/recuperación para navegador y pruebas de restauración central deben demostrar qué se recupera y qué puede perderse. Cifrar/proteger copias según los datos que contengan. El piloto documenta quién actúa ante cada incidente y cómo contactar soporte.
