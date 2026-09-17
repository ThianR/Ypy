# Contratos de aplicación y POS

Estos contratos son propuestas de implementación. Se formalizan como OpenAPI, esquemas de eventos y fixtures durante E02/F02; no afirman que existan endpoints hoy.

## API central

Prefijo `/api/v1`. Operaciones propuestas: sesión y contexto, catálogos por versión, apertura/cierre de caja, recepción/compra, confirmación de venta, cobro/pago, bloques de numeración, configuración firmada de terminal, integración de operaciones y consulta de conflictos. Usar comandos específicos (`confirmar`, `devolver`, `anular`), no CRUD que permita modificar directamente documentos confirmados.

Cada comando incluye clave de idempotencia. En la integración POS `operation_id` corresponde al `uid` globalmente único de V4; verificar además empresa, terminal, tipo y contenido antes de devolver el resultado. Para otros comandos, documentar su ámbito de unicidad con restricciones persistentes. Un UUID repetido con el mismo contenido semántico devuelve el resultado original; con otro contenido es conflicto. Definir serialización canónica de decimales/campos antes de calcular hashes. Identidad y empresa se contrastan contra la sesión; no bastan los identificadores del JSON.

Errores tienen código estable, mensaje comprensible, campos afectados, correlación y condición de reintento. Diferenciar error de validación, falta de autorización, conflicto y falla temporal. El usuario puede inspeccionar una operación rechazada; el cliente no inventa un UUID nuevo para forzar su aceptación.

## Operación local

Un sobre versionado conserva:

- UUID de operación, empresa, sucursal, terminal y sesión de caja.
- Secuencia creciente por terminal y versión de contrato/aplicación.
- Fecha comercial, zona horaria y fechas de ocurrencia local, recepción e integración central separadas; conservar política/version de día comercial según [07](07_OPERACION_Y_RECUPERACION.md).
- Usuario y referencia verificable del permiso offline firmado.
- Tipo documental, talonario, bloque y número asignado, cuando corresponda.
- Líneas con identificador de artículo/presentación, cantidad base, lotes/series, precio, descuento, impuestos y versiones de catálogo usadas.
- Moneda, totales, cotizaciones aplicadas y política de redondeo; costo local provisional separado del central definitivo.
- Cobros por medio y moneda, cambio entregado y aplicaciones, sin almacenar secretos de tarjetas.

En ambos adaptadores, guardar venta+cobro+consumo de número+saldo provisional+cola en una única transacción local. Preparar los datos antes de iniciar la transacción; ninguna llamada de red o impresión forma parte de ella. En el cliente del agente, el comando completo cruza la API local una sola vez; no encadenar endpoints independientes para vender, numerar y cobrar.

La venta local confirmada no se reescribe al sincronizar. Su reconocimiento central y eventual conflicto son estados adicionales. Operación local válida no garantiza aceptación fiscal ni autorización de un adquirente.

## Sincronización y orden

Estados sugeridos de cola: PENDIENTE, ENVIANDO, APLICADA, CONFLICTO. ENVIANDO es recuperable tras reinicio. Reintentar con backoff y jitter; sincronizar al arrancar, mientras la aplicación esté activa, al recuperar conexión y por acción manual. No confiar únicamente en detección de conectividad ni Background Sync.

Procesar operaciones relacionadas en orden por terminal/caja; el servidor detecta huecos y duplicados. Un conflicto que afecta la secuencia requiere resolución visible antes de aplicar dependencias. No existe orden cronológico global garantizado entre cajas aisladas; el costo central sigue el orden de integración aprobado en V4.

El acuse incluye UUID, resultado, identificadores centrales, fecha de integración y versión/cursor de sincronización. Responder éxito únicamente después del commit. Pérdida del acuse provoca reenvío idempotente. No eliminar silenciosamente cola ni registros auditables tras el ACK; establecer retención y respaldo.

Los snapshots de saldo/catálogo llevan versión/cursor y cobertura de operaciones. Al aplicarlos, conservar pendientes y no restar otra vez operaciones ya incluidas. La referencia V4 debe transformarse en fixtures de ambos adaptadores. Una actualización de catálogo no cambia precio/impuesto de una venta ya confirmada.

## Numeración y permisos

Bloques exclusivos por terminal, principal y reserva. Reabastecer al llegar al umbral configurado mientras haya conectividad. Al agotar ambos, usar solo alternativa prehabilitada compatible; sin ella, conservar borrador pendiente sin emitir ni inventar número.

Permisos offline con firma, versión, empresa/usuario/terminal, alcance y vencimiento. Mantener regla inicial V4 de 72 horas y extensión única de supervisor hasta 24 horas, configurable según política documentada. La revocación central no puede conocerse instantáneamente estando aislado; el vencimiento limita esa ventana. Registrar retrocesos de reloj e impedir que reiniciar renueve permisos. Un equipo con control administrativo total no ofrece una garantía de reloj inviolable.

No restaurar una copia antigua y reutilizar sus números como si fuera la terminal actual: requerir reconciliación, nueva identidad/rangos si corresponde y preservación de la trazabilidad. No transferir rangos automáticamente entre modalidades ni liberar números por timeout.

## Importes, cantidades y conversión

- API: identificadores BIGINT como cadenas; UUID como cadenas; importes, cantidades y tipos de cambio como cadenas decimales canónicas. Nada de float64/Number para aritmética financiera.
- Cantidad base: precisión V4 de seis decimales. Costos y factores usan la precisión interna correspondiente del esquema; no reducirlos a decimales de moneda antes de calcular.
- Moneda de cada operación y moneda base explícitas. Ejemplo de convención: `importe_destino = importe_origen × tasa_origen_destino`; registrar origen/destino, tipo de cotización, fecha y tasa usada. La inversa y los cruces tienen reglas explícitas, nunca una selección oculta.
- Configuración de decimales por moneda; redondeo propuesto comercial a mitad alejándose de cero, salvo regla fiscal aplicable. F02/R01 valida casos de impuesto, descuento, devolución y sumatoria; no cambiar política en documentos históricos.
- Guardar importe original y equivalente base al confirmar; para pagos multimoneda conservar importe entregado, cambio y valor aplicado en cada moneda relevante. Diferencias de cambio/redondeo tienen tratamiento separado y trazable.
- Consulta en otra moneda usa un criterio visible: cotización histórica de operación o cotización de reporte a una fecha. No confundir ambos ni alterar el comprobante original. Si falta cotización, indicar no disponible o bloquear conversión; nunca asumir tasa 1.
- Offline solo usa monedas/cotizaciones descargadas y vigentes según la política de terminal. E02/F02 inventaría campos existentes y propone migraciones para cualquier dato que falte; no suponer que todas las tablas V4 ya conservan esos detalles.

## Impresión y compatibilidad

Generar ticket desde una venta durable y su estado fiscal correspondiente. Cola de impresión independiente, con estado desconocido si se pierde respuesta de la impresora. Reimpresión explícita y trazable. Modo kiosco es presentación/configuración; la ruta sin instalación no promete corte/cajón/impresión silenciosa en todos los equipos.

Versionar API central, contrato local y esquema de almacenamiento por separado. Propuesta inicial: servidor acepta versión actual y anterior del contrato POS durante una ventana de actualización acordada. No migrar esquema local mientras otra ventana vende; coordinar acceso y cerrar limpiamente. No activar actualización de PWA en medio de una operación. Un cliente incompatible recibe instrucciones y conserva pendientes para migración, sin borrarlos.

## Contratos adicionales de negocio

La [especificación de sincronización](05_SINCRONIZACION_POS.md) precisa autoridad, UUID global, orden, resoluciones de conflicto y bootstrap. [Seguridad](06_SEGURIDAD_Y_AUDITORIA.md) define ámbito y auditoría atómica.

Caja: apertura/fondo, ingreso manual, retiro, arqueo y cierre son comandos trazables ligados a una sesión. Ingreso/retiro no crean automáticamente venta, deuda ni pago a proveedor. Se integran idempotentemente también offline; cierre definitivo exige resolver operaciones anteriores. La suma esperada distingue moneda/medio y cambio entregado para no contarlos dos veces.

Conteo físico conectado: para el MVP, bloquear temporalmente los movimientos del alcance contado solo si puede garantizarse que no quedan cajas offline vendiendo sobre él; en otro caso bloquear la confirmación del ajuste hasta reconciliar. Guardar saldo de referencia, conteo, diferencia, alcance/lote y motivo; confirmar una sola vez mediante ajustes V4. No restar simplemente el saldo actual a un conteo antiguo mientras siguen entrando ventas. El conteo concurrente avanzado queda fuera del primer circuito.

Código de barras: conservar código leído y regla aplicada. Resolver código normal por presentación o peso/precio por `gp_regla_balanza`, con formato y decimales configurados. Definir prioridad ante coincidencia exacta y prefijo para evitar ambigüedad; rechazar formatos ambiguos. Para precio incorporado, precisar si el importe incluye impuestos y cómo se obtiene cantidad con el precio vigente fijado en la operación; rechazar división por cero o cantidad no representable según política. No confundir lectura de etiqueta con integración directa a una balanza: esta última requiere protocolo/hardware específico.
