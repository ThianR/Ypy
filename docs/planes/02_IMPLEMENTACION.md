# Backlog de implementación

Todas las tareas empiezan pendientes. Cada entrega requiere código, migración si aplica, prueba de aceptación y documentación operativa. Responsable lógico es el módulo indicado; asignar personas al planificar trabajo. No hay fechas estimadas hasta conocer el equipo.

## E — preparación y validación técnica

| ID | Depende de | Trabajo y entregable | Criterio de aceptación |
|---|---|---|---|
| E00 | Proyecto nuevo Ypy confirmado | Crear el esqueleto de aplicación en Ypy; registrar alcance inicial y verificar instrucciones locales | Ruta y proyecto nuevo documentados; compilación mínima reproducible |
| E01 | E00 | Arquitectura: manifiesto de tablas/funciones V4 por propietario, dependencias y excepciones | Cada objeto tiene propietario; funciones cruzadas enumeradas; regla CI impide importar internos de otro módulo |
| E02 | E00 | Plataforma: Go/Vue/PrimeVue, controladores SQL, librerías decimales y ejecutor de migraciones; fijar versiones compatibles | API y frontend compilan; llamada real a función V4; baseline y catálogos instalan sin datos de prueba |
| E03 | E02 | POS: prototipo desechable o reutilizable mínimo Dexie y SQLite; contrato común; impresión simulada | Ambas rutas guardan operación+número+cola, sobreviven reinicio y reciben reenvío sin duplicarlo |

E03 prueba mecanismos con operaciones de laboratorio; todavía no certifica ventas ni emisión fiscal. Probar periféricos reales tan pronto estén disponibles.

## F — fundamentos del producto

| ID | Depende de | Trabajo y entregable | Criterio de aceptación |
|---|---|---|---|
| F01 | E01, E02 | Identidad: sesión, autorización empresa/sucursal, roles de aplicación y migraciones, alta controlada de terminal | Cambiar `empresa_id` en una solicitud no permite acceder a otra empresa; usuario inactivo no inicia sesión |
| F02 | E02 | Maestros: tipos monetarios, conversión/rounding y contrato JSON; fixtures Go/TS/SQL | Igual resultado en los tres entornos; BIGINT y decimales no pierden precisión; no cambiar valor histórico al cambiar cotización |
| F03 | F01, F02 | Maestros/catálogo: artículos, barras, presentaciones, terceros, precios, monedas/impuestos; CRUD mínimo y buscador POS | Se resuelve código de barras y unidad; impuesto/precio versionados; artículo inactivo no se ofrece para nuevas ventas |
| F04 | F01 | Numeración/POS: talonarios públicos por sucursal o asignados, terminales, bloques principal/reserva y permisos firmados | Dos cajas reciben rangos exclusivos; no hay MAX+1; terminal no utiliza rango ajeno; vencimiento de permiso comprobado |
| F05 | F01, E01 | Plataforma: logs con request_id/correlation_id, errores API, completar `gs_evento_salida` y recepción durable `gp_evento_entrada`, diagnóstico y procesamiento recuperable | Evento solo aparece después del commit; reintento no duplica efecto; logs no exponen secretos; no duplicar tablas equivalentes |
| F06 | F01, F05 | Auditoría funcional con `gs_auditoria`, ámbitos caja/depósito y configuración gradual versionada; revisar extensiones necesarias | Cambio sensible exige permiso/motivo; auditoría confirmada atómica; intentos rechazados conservan evidencia separada; flag no concede permiso |

## V — primer circuito comercial online

| ID | Depende de | Trabajo y entregable | Criterio de aceptación |
|---|---|---|---|
| V01 | F02, F03 | Compras/inventario: recepción mínima y consulta de saldo/costo/kardex | Recibir 10 unidades suma una vez; repetir confirmación no suma; consulta de saldo usa estado actual |
| V02 | F01, F02, F04 | Tesorería: apertura, cobro por medios y cierre básico | Apertura por usuario/caja autorizados; medios suman total según redondeo; diferencia de cierre identificada |
| V03 | V01, V02, F05, F06 | Workflow venta contado: borrador, confirmación, cobro, stock y número atómicos | Fallo de cobro no deja venta/stock confirmado; reintento retorna mismo resultado; venta no POS sin saldo se bloquea |
| V04 | V03 | Frontend POS: búsqueda/lector, cantidades, totales, medio de pago, confirmación y ticket | Venta completa sin navegar por pantallas administrativas; doble clic no crea otra venta; sin papel se reimprime la misma |

## P — POS aislado en ambas modalidades

| ID | Depende de | Trabajo y entregable | Criterio de aceptación |
|---|---|---|---|
| P01 | E03, V03, F04 | Adaptador Dexie/PWA: catálogo local, permisos, caja, venta+cobro+número+stock provisional+cola en transacción | Arranque offline tras preparación, venta y reinicio conservan datos; segunda ventana no reutiliza número; manejo visible de cuota insuficiente |
| P02 | E03, V03, F04 | Agente Go/SQLite: mismo contrato, interfaz local y proceso de impresión | Mismos fixtures que P01; reiniciar agente recupera operación; ningún acceso remoto/no autorizado controla la caja |
| P03 | P01, P02, F05 | Integración central completa: adaptar V4 para venta+cobro+acuse y conflictos sin efectos parciales | Reenvío o pérdida de respuesta produce una venta y un cobro; UUID con contenido distinto se rechaza; integración tardía conserva fecha comercial |
| P04 | P03 | Sincronización en ambas rutas: autoridad de datos, bootstrap/versiones, colas, orden, snapshots, reintentos y resolución de conflictos; estado visible | ACK no descuenta de nuevo; snapshot conserva pendientes; huecos esperan antecedentes; resolución auditada permite continuar; cajero ve pendientes/último éxito |
| P05 | P04 | Multimoneda offline y medios de pago; apertura/cierre local con integración ordenada | Usa cotizaciones descargadas y fijadas; no inventa cambio faltante; cierre local provisional se concilia tras integración de todas sus ventas/cobros |
| P06 | P04 | Recuperación, actualización y periféricos reales | Actualizar conserva cola/rangos; restauración no reutiliza numeración consumida; ticket probado en modelos declarados compatibles |
| P07 | P03 | Simulador POS parametrizable y reproducible: terminales, operaciones, duplicados, huecos, fallos y reinicios | T25 reproduce fallo por semilla; llegada desordenada no viola secuencia; tras resolución/reenvío cada operación válida tiene un solo efecto |

P01 y P02 pueden implementarse sucesivamente, reutilizando interfaz y fixtures. La modalidad que esté lista primero permite demostraciones, pero no completa por sí sola el requisito de ambas.

## C — compras, inventario y operación completa

| ID | Depende de | Trabajo y entregable | Criterio de aceptación |
|---|---|---|---|
| C01 | V01, V03 | Órdenes, factura, recepción previa y compra directa con recepción automática | 3 recepciones parciales y una factura producen solo el ingreso recibido; vínculos no exceden cantidades; factura anterior también funciona |
| C02 | C01, F02 | Diferencias de precio, costo negativo y costo pendiente; pantallas de trazabilidad | Diferencia distribuida conforme V4; no recalcular desde saldo inicial congelado; costo pendiente visible hasta ajuste autorizado |
| C03 | C01, F03 | Lotes, vencimientos, series y conversiones de presentación | No duplicar serie; unidades base correctas; venta exige trazabilidad configurada; política de vencimientos definida y probada |
| C04 | V03, C01 | Devoluciones, notas de crédito y anulaciones; inicialmente gestión central conectada | Devolución parcial respeta acumulado y costo original; nota financiera no duplica devolución física; bloqueo por dependencias explicado al usuario |
| C05 | V02, C01, P05 | Cuentas por cobrar/pagar, pagos y aplicación de créditos; cierres completos | Saldos pendientes y medios cuadran con comprobantes; pago mixto/multimoneda conserva aplicaciones y cambios |
| C06 | C03, V03 | Transferencias, preparación/despacho y entrega básicos | Origen/destino y kardex concilian; despacho de venta confirmada no resta stock nuevamente |
| C07 | V01, C03, F06 | Conteo físico conectado, saldo de referencia, aprobación y ajuste usando `gi_ajuste_*`; agregar documento de conteo si falta | T27: diferencia autorizada una sola vez; alcance contado no cambia sin control; pendientes offline impiden confirmación insegura |
| C08 | F03, F02, V04 | Lectura de códigos de peso/precio incorporado con `gp_regla_balanza`, reglas descargadas en ambas modalidades | T28: mismo resultado online/offline; prioridad exacto/prefijo explícita; formato ambiguo o precio cero no produce cantidad inválida |
| C09 | V02, P05, F06 | Caja: fondo, ingresos/retiros con motivo, arqueo por moneda/medio y cierre; ampliar V4 si faltan movimientos propios | T29: no genera ventas/deudas ficticias; cada movimiento ligado a sesión; reenvío offline único; esperado concilia incluyendo cambio |

Fuera del circuito offline inicial: devoluciones/notas, ajustes de inventario y compras desde caja aislada. Permanecen disponibles conectadas; agregarlas offline requiere ampliar contratos y pruebas. La venta/cobro/caja siguen funcionando aisladas como requisito principal.

## R — fiscal, reportes y piloto

| ID | Depende de | Trabajo y entregable | Criterio de aceptación |
|---|---|---|---|
| R01 | F04, F05, empresa piloto | Fiscal: verificar modalidad/habilitación vigente, flujo de emisión y contingencia, implementación del adaptador elegido | Evidencia del flujo fiscal y representación requerida; pruebas en entorno autorizado; ningún documento pendiente se presenta como aceptado |
| R02 | C02, C05, C06 | Reportes: ventas/caja, compras/deuda, stock/kardex y costo; PDF, Excel y gráficos básicos | Totales cuadran con el origen; conversiones indican moneda, fecha y criterio; exportación respeta empresa/permisos |
| R03 | P06, P07, F06, C01–C09, R01, R02 | Matriz completa de fallos/carga, empaquetado, respaldo y guía de soporte | Casos críticos de 04 aprobados; objetivos de rendimiento acordados y medidos; instalación y restauración demostradas |
| R04 | R03 | Piloto acotado con cajas y usuarios definidos | Cierre conciliado y pendientes visibles; salida/reversión operativa ensayada; responsable del negocio acepta los circuitos reales |

Diseñar R01 desde F04; su dependencia de una empresa piloto no detiene fundamentos ni pruebas internas. Todas las modalidades fiscales no constituyen requisito de la primera entrega: sí una arquitectura extensible y una vía legal operativa para el piloto.

## Secuencia para demostrar valor temprano

E00/E01/E02 → fundamentos mínimos F01/F02/F03/F04/F05/F06 → V01/V02/V03/V04 → P01/P02/P03/P04. F03 arranca con un catálogo mínimo y una recepción de prueba; no esperar a C01 (compras completas) para probar venta offline. E03 ensaya persistencia mientras se construyen fundamentos. P07 empieza en cuanto el protocolo de integración sea ejecutable. F02/F04 incluyen fecha comercial y política de turno de 07, aceptadas en T30.

Ampliar después los catálogos y circuitos C, incluidos C07–C09. Los contratos de 05–07 son parte de los entregables, no documentos optativos de soporte. R02 debe actualizarse con los nuevos movimientos de caja y conteos antes de R03 aunque sus primeros reportes puedan desarrollarse antes.

## Regla de finalización de tarea

Una tarea está completa cuando tiene entrega demostrable, aceptación ejecutada, migración reproducible si corresponde, errores recuperables y límites documentados. Actualizar aquí el estado y enlace a evidencia/commit al terminar. Una tabla creada, una pantalla dibujada o una prueba simulada no sustituyen la aceptación integral correspondiente.

