# Validación, despliegue y piloto

## Evidencia inicial

V4 entrega pruebas de SQL, concurrencia, instalación y una referencia Node/SQLite. Conservarlas y ejecutarlas contra el baseline durante E02. No equivalen a pruebas de API, interfaz, Dexie, agente Go, cobros offline ni emisión fiscal. Las nuevas pruebas se añaden por comportamiento y reutilizan casos de negocio compartidos.

## Matriz mínima de aceptación

| ID | Escenario | Resultado obligatorio | Nivel |
|---|---|---|---|
| T01 | Instalación nueva y siguiente migración | Estructura/catálogos válidos; sin datos de laboratorio en producción; versiones registradas | PostgreSQL real |
| T02 | Usuario/terminal intenta otra empresa o sucursal no autorizada | Acceso denegado sin filtrar documentos ajenos | API |
| T03 | Dos consumidores asignan números y bloques concurrentemente | Sin superposición ni reutilización; UUID repetido devuelve mismo resultado | PostgreSQL/API |
| T04 | Recepción 10+20+20, con ventas intercaladas | Cada recepción usa saldo/costo vigentes; ingreso total 50; factura no ingresa otras 50 | Integración |
| T05 | Factura anterior a recepción y diferencia de precio posterior | Aplicación parcial correcta; capitalización/diferencia consumida según V4 | Integración |
| T06 | Venta no POS sin saldo, y POS autorizado sin saldo | Primera bloqueada; segunda registra negativo controlado y costo o pendiente explícito | API y ambos POS |
| T07 | Confirmación con falla al registrar cobro | Sin venta/stock/cobro parcialmente aplicados | Integración con inyección de fallo |
| T08 | Doble clic, timeout y reenvío de UUID | Una venta, un número, un movimiento y un cobro; contenido distinto provoca conflicto | API y ambos POS |
| T09 | Caída después de commit central antes del ACK | Reenvío devuelve resultado original; cliente no descuenta otra vez | Integración y ambos POS |
| T10 | Caja aislada, cierre de app y reinicio del equipo/proceso | Ventas, caja, cola y rangos conservados; se puede continuar conforme permisos | Ambos POS en entorno objetivo |
| T11 | Dos ventanas de una terminal venden a la vez | Exclusión de sesión o serialización transaccional; números únicos y saldos consistentes | Ambos POS |
| T12 | Permiso vencido, extensión usada y retroceso de reloj | Política aplicada; no extensión ilimitada por reinicio | Ambos POS |
| T13 | Bloque principal y reserva agotados | Alternativa compatible o pendiente sin emisión; no se inventa número | Ambos POS |
| T14 | Dos cajas aisladas, ventas tardías y compras centrales | Sin duplicados; costos centrales por integración; fecha/costo locales preservados | Integración de extremo a extremo |
| T15 | Snapshot llega con operaciones pendientes o ya aplicadas sin ACK local | Saldo no pierde pendientes ni duplica descuentos | Ambos adaptadores |
| T16 | Pago mixto, redondeo, devolución y varias monedas | Go/TS/SQL coinciden; original y cambio históricos preservados; falta de cambio visible | Fixtures e integración |
| T17 | Devolución parcial repetida o anulación con dependencias | Límite acumulado respetado; reversión trazable; no doble stock | Integración |
| T18 | Impresora sin papel o respuesta incierta | Venta permanece; reimpresión identificada sin nuevo comprobante comercial | Equipo real |
| T19 | Transferencia y despacho de venta confirmada | Saldos origen/destino correctos; despacho no descuenta nuevamente | Integración |
| T20 | Actualización con cola pendiente, pestaña antigua y agente anterior | Datos conservados; versión incompatible no corrompe operaciones | Ambos POS y servidor |
| T21 | Cuota de navegador insuficiente / disco lleno | Falla transaccional visible; no informar venta guardada si no persistió | Ambos POS |
| T22 | Restauración de servidor/caja | Procedimiento detecta posibles acuses/rangos posteriores al respaldo; no acepta reutilización de número | Simulacro operativo |
| T23 | Reintento de envío fiscal, rechazo y respuesta tardía | Estados coherentes con modalidad; no mostrar pendiente como aceptado; trazabilidad | Adaptador fiscal |
| T24 | Reporte, cierre y saldos pendientes | Totales concilian; moneda y criterio de conversión visibles; aislamiento por empresa | Integración/UI |
| T25 | Caos de sincronización: enviar B,D antes de A,C; perder respuestas, reiniciar y reenviar | Se detectan huecos sin efectos prematuros; tras recibir antecedentes cada operación válida y sus cobros se aplican una sola vez | Simulador P07 y ambos adaptadores |
| T26 | Cambio sensible, ámbito indebido y resolución de conflicto | Permiso/motivo obligatorios; evidencia atómica; rechazados registrados aparte; ningún flag elude autorización | F01/F06, API |
| T27 | Conteo físico con movimientos concurrentes o ventas offline pendientes | Ajuste no se confirma contra una referencia inválida; diferencia válida se aplica una sola vez y conserva motivo | C07, integración |
| T28 | Código normal, peso/precio incorporado y prefijos ambiguos | Cantidad/unidad/impuesto correctos; misma regla offline; entradas ambiguas o imposibles rechazadas | C08, fixtures y ambos POS |
| T29 | Fondo, ingreso, retiro, cobro con cambio y cierre en varias monedas | Esperado concilia por medio/moneda; sin ventas ficticias; reenvíos no duplican movimientos | C09, integración y ambos POS |
| T30 | Turno cruza medianoche; cambia zona/corte o se atrasa reloj local | Fecha comercial y política originales conservadas; recepción/procesamiento separados; secuencia no depende del reloj | F02/F04, ambos POS y reportes |

La pérdida total del almacenamiento local antes de sincronizar puede perder operaciones; ninguna prueba permite prometer recuperación sin una copia disponible. Documentar frecuencia de respaldo y pérdida máxima aceptable por modalidad. Borrar el sitio del navegador requiere tratamiento específico y capacitación operativa.

T25 usa cuatro ventas de una línea stock cada una para esperar exactamente cuatro salidas; no generalizar ese conteo de movimientos a ventas multilínea o sin stock. Añadir variante con B en conflicto: no ejecutar C/D antes de resolución auditada; luego confirmar la secuencia sin reescribir B. Guardar semilla, configuración, trazas y conciliación final. Un UUID de otra empresa no devuelve datos previos aunque coincida globalmente.

Extender T20/T22 con cambio de instalación física y credenciales: la instalación reemplazada no obtiene nuevos rangos; documentar ventana de permisos offline aún vigentes y conciliar consumos anteriores. La clonación deliberada de un equipo no se considera resuelta solo con UUID: probar detección de colisiones y procedimiento de incidente.

## Capas de pruebas y CI

1. Pruebas unitarias para cálculo, estados y reglas independientes de SQL; fixtures compartidos para Go/TypeScript.
2. Integración en PostgreSQL real para funciones V4, API y transacciones. No sustituirlas por SQLite.
3. Pruebas contractuales de los dos adaptadores con los mismos casos de venta, permisos, números y cola.
4. Pruebas de interfaz completas para los caminos principales y fallos de conexión/reinicio.
5. Impresión, periféricos y cortes del entorno real en laboratorio del piloto.

CI ejecuta formato/compilación, tipos frontend, límites de módulos, pruebas relevantes y migraciones desde base vacía. Un cambio de costo/stock/numeración/sincronización requiere la matriz afectada y las pruebas de concurrencia. Cada evidencia indica commit, versiones, entorno y resultado.

## Rendimiento

Antes de aprobar piloto, acordar número de cajas, catálogo, líneas por venta, ventas por minuto, backlog tras desconexión y objetivo de latencia p95. Medir venta normal, recepción concurrente, recuperación de varias cajas y generación de reportes. Registrar tiempo de espera por bloqueo de empresa, conexiones, CPU/memoria y tamaño de cola.

La estrategia V4 serializa ciertas operaciones por empresa. Si incumple el objetivo, revisar alcance/duración de transacciones y luego bloqueo por artículo/talonario con orden consistente; repetir pruebas de integridad. No declarar capacidad de supermercado por haber pasado ocho conexiones de prueba.

## Despliegue y operación

- Configuración por entorno, secretos externos al repositorio y usuario de aplicación distinto al de migraciones. No reutilizar el PostgreSQL portátil de laboratorio con autenticación de confianza en producción.
- Servidor: artefacto versionado, frontend compilado, conexión PostgreSQL, HTTPS y supervisión del proceso. Definir si será servidor local, alojado o ambos al seleccionar piloto.
- Reportes: Chrome/Chromium cuando aplique, cupo de trabajos, tiempo máximo, almacenamiento/retención y acceso autorizado a descargas.
- Agente: instalador/inicio automático, permisos mínimos, identidad por caja, SQLite con configuración de durabilidad ensayada y actualización controlada.
- Navegador: origen estable, preparación inicial online, política de persistencia, actualización segura de PWA y procedimiento de recuperación. HTTPS salvo entorno local admitido por el navegador.
- Copias PostgreSQL y locales consistentes; ensayar restauración. Definir RPO (pérdida máxima aceptable) y RTO (tiempo máximo de recuperación) con el negocio.
- Un respaldo del servidor anterior a ventas ya reconocidas requiere conciliación de journal local/central. No purgar evidencias locales hasta cumplir retención y recuperación definidas. Una copia de archivos SQLite activos sin coordinación no constituye por sí sola un respaldo válido.
- Monitorear fallos, conflictos, antigüedad de pendientes, rangos próximos a agotarse, permisos, jobs fiscales y conciliación. Alertas deben indicar la acción operativa necesaria.

## Condiciones para iniciar piloto real

Ambas modalidades cumplen los escenarios aplicables; inventario/caja/cuentas concilian; periféricos declarados funcionan; modalidad fiscal y habilitación del contribuyente están verificadas; respaldo/restauración ensayados; límites de capacidad medidos y conocidos.

Definir usuarios, cajas y alcance acotado, saldo inicial aprobado, horario de soporte y procedimiento ante fallos. No regresar a una base anterior descartando ventas ya emitidas: detener nuevas operaciones si hace falta, preservar evidencias y conciliar antes de cambiar de versión o sistema.

Durante el piloto registrar diferencias y pendientes por turno. Ampliar alcance cuando los cierres concilien y se hayan resuelto incidentes críticos. La aceptación final del negocio se realiza sobre circuitos demostrados, no sobre cantidad de tablas o pantallas.
