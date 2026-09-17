# Sincronización POS: autoridad, recepción y recuperación

Complementa el contrato de operaciones de [03_CONTRATOS.md](03_CONTRATOS.md). Aplica a ambas modalidades; no cambia la política fiscal, los bloques de numeración ni las reglas de costos de V4.

## Autoridad de los datos

| Información | Autoridad | Copia local y conciliación |
|---|---|---|
| Empresa, sucursal, usuario y artículo | Central | Caché versionada; nuevas modificaciones solo conectadas |
| Lista de precios, impuestos, unidades y cotizaciones | Central para configuración nueva | La operación confirmada conserva su snapshot; descargar cambios no reescribe ventas |
| Permisos y configuración de terminal | Central | Snapshot firmado y acotado; no se puede conocer una revocación mientras la terminal está aislada |
| Venta/cobro/movimiento de caja creado offline | Terminal de origen para el registro original | Central valida e integra; conserva el original también en conflicto. Un rechazo no autoriza borrar ni cambiar lo que ocurrió |
| Stock mostrado offline | Estimación local | Snapshot central más operaciones no incluidas; no representa disponibilidad global garantizada |
| Kardex, saldo global y costo definitivo | Central | El costo local es provisional y se conserva separado del resultado de integración |
| Rango de números | Central asigna; terminal consume su rango exclusivo | Conciliar consumos; no reasignar por timeout ni inventar números fiscales |
| Estado fiscal | Adaptador central conserva evidencia de la autoridad fiscal cuando aplique | Una venta local o un ACK comercial no equivalen a aceptación fiscal |

Los desacuerdos se resuelven según esta tabla y con operaciones de corrección trazables, no con una regla genérica de «gana el último cambio».

## Aprovechar V4

`gp_evento_entrada` ya representa la recepción durable central: UUID único global, secuencia única por empresa/terminal, fechas de ocurrencia y recepción, contenido y estado. `gs_evento_salida` ya ofrece la base de outbox. Reutilizar/extender estos objetos mediante migraciones: no crear un segundo inbox/outbox equivalente. Inventariar qué funciones los escriben y completar deduplicación, resultado previo, hash canónico, fecha de procesamiento y reintentos donde falten.

`operation_id` en el protocolo corresponde al UUID `uid` existente, sin renombrar automáticamente columnas. Conservar su unicidad global y comprobar empresa/terminal/tipo/contenido antes de devolver cualquier resultado; una colisión de UUID no debe filtrar información de otra empresa. La secuencia también detecta dos UUID diferentes intentando ocupar la misma posición.

Una terminal envía al menos una vez; el servidor aplica el efecto comercial una sola vez mediante transacción y deduplicación persistente. El resultado y ACK se obtienen después del commit. Una recepción pendiente o en conflicto no debe bloquear para siempre un reintento válido por el solo hecho de que el UUID ya exista.

## Estados y orden

Conservar en local PENDIENTE, ENVIANDO, APLICADA y CONFLICTO. Agregar un motivo estructurado y una acción requerida; no crear seis estados redundantes. Errores temporales vuelven a pendiente; errores de negocio quedan visibles para resolución. ENVIANDO es recuperable después de reiniciar.

Cada secuencia por terminal tiene una posición de resolución. Un hueco retiene las operaciones posteriores o devuelve `SYNC_SEQUENCE_GAP` sin efectos. Un reenvío fuera de orden no se considera por sí solo rechazo comercial definitivo. Registrar el tratamiento elegido en P03 y probarlo en ambos clientes.

Un conflicto que no se puede aplicar se resuelve por un comando de supervisor auditado: reconocer la incidencia y, según el caso, asociar una compensación o resolución explícita sin efectos originales. Solo entonces puede avanzar la secuencia. No editar el payload original, saltar huecos en silencio ni reutilizar un número. El recibo de resolución del servidor permite a la terminal continuar sin reintentar indefinidamente la misma operación. Casos fiscales requieren su procedimiento específico.

## Identidad, bootstrap y versiones

Separar terminal lógica (`gp_terminal`), instalación/dispositivo y sesión de caja (`gf_caja_apertura`). El reemplazo de PC necesita enrolamiento y revocación de la instalación anterior, conciliación de pendientes y nuevos rangos cuando corresponda. No clonar identidad/rangos copiando SQLite a dos equipos. E01/F04 revisan qué metadatos faltan en V4 antes de proponer migraciones.

Bootstrap entrega versiones de catálogo, precios, permisos y configuración, modalidad de terminal y capacidades disponibles. Descarga inicial completa; cambios posteriores con cursor, borrados/inactivaciones explícitos y aplicación atómica. Si un cursor expiró, reconstruir caché desde snapshot manteniendo journal pendiente. No exigir incrementales sofisticados para la primera venta de laboratorio, pero sí definir el contrato antes del piloto.

Contrato POS actual/anterior y versión del esquema local son conceptos distintos. Compatibilidad se prueba por versión de payload, no comparando solo números de versión del ejecutable. Actualizar no debe invalidar ventas ya guardadas que estuvieron desconectadas.

## Estado visible y simulador

La caja muestra conexión, pendientes, conflictos y última sincronización exitosa. Soporte ve versión de aplicación/esquema, terminal, antigüedad de pendientes y rangos restantes, sin exponer secretos ni detalles técnicos innecesarios al cajero.

P07 implementa un simulador parametrizable: número de terminales, operaciones, semilla y fallos. Probar duplicados, huecos, timeouts después de commit, reinicios y recuperación. Conservar semilla y resultados para reproducir un fallo. Una escala de 10 cajas o miles de ventas es un escenario de ensayo, no una promesa de capacidad. El simulador del protocolo no reemplaza pruebas reales de IndexedDB, SQLite ni periféricos.
