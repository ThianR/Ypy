# Ypy — planes de implementación del MVP

Proyecto confirmado: nuevo, ubicado en `D:\Personales\SISTEMAS\Genesis\Ypy`. Nombre del producto: **Ypy**. Estos son los planes vigentes; Venture no forma parte de la implementación.

## Resultado esperado

ERP como monolito modular Go/PostgreSQL, interfaz Vue 3 + TypeScript + PrimeVue y dos modalidades POS: navegador con Dexie/PWA y servicio local Go con SQLite. Primer mercado: minimercados/supermercados en Paraguay. Mantener la extensibilidad del modelo V4 y la nomenclatura actual (`empresa_id`).

Este paquete define trabajo ejecutable; no acredita código, pruebas ni despliegues realizados. Estado inicial de todas las tareas: pendiente. La base es `creacionV4.sql`, sus catálogos, contratos y validaciones existentes en el directorio Genesis, tres niveles por encima de estos planes. E02 incorporará una copia versionada y verificada al proyecto; todavía no hay migraciones instaladas.

## Documentos de trabajo

1. [Arquitectura y límites](01_ARQUITECTURA.md): propietarios, dependencias, repositorio, transacciones y futura extracción.
2. [Implementación por entregas](02_IMPLEMENTACION.md): tareas identificadas, dependencias, entregables y aceptación.
3. [Contratos POS y multimoneda](03_CONTRATOS.md): operaciones, sincronización, importes y compatibilidad.
4. [Pruebas y salida a piloto](04_VALIDACION_Y_PILOTO.md): escenarios compartidos, despliegue y recuperación.
5. [Sincronización POS](05_SINCRONIZACION_POS.md): autoridad de datos, inbox/outbox existentes, secuencias, conflictos e identidad de instalación.
6. [Seguridad y auditoría](06_SEGURIDAD_Y_AUDITORIA.md): permisos por ámbito, evidencia funcional y habilitaciones graduales.
7. [Operación y recuperación](07_OPERACION_Y_RECUPERACION.md): diagnóstico, fecha comercial, actualizaciones y procedimientos por incidente.
8. [Interfaz, formularios y personalización](08_INTERFAZ_FORMULARIOS_Y_PERSONALIZACION.md): componentes comunes, tablas, filtros, campos tipados, adaptación a dispositivos y preferencias de colores y vistas por usuario. Plan propuesto; implementación pendiente. Para este trabajo, la base propuesta es Vue con componentes comunes propios; la mención inicial a PrimeVue no supone que esté instalado ni exige incorporarlo.

## Decisiones de partida

- Un backend central desplegable, organizado en módulos. PostgreSQL compartido inicialmente. El agente local POS es un componente de borde, no un microservicio del ERP.
- Una interfaz POS común y dos adaptadores de persistencia. Cada terminal elige una modalidad; no cambia automáticamente al fallar una.
- Mantener reglas V4: costo promedio móvil empresa/artículo por recepción, ventas no POS bloqueadas ante insuficiencia, negativo POS controlado, rangos preventivos y reserva, acuses sin duplicación y kardex valorizado.
- Impuestos, importes, monedas y cambios se conservan en cada operación; las conversiones de consulta no reescriben documentos confirmados.
- Conservar las funciones transaccionales existentes como autoridad central de stock/costo/numeración. No copiar ese núcleo a Go durante el arranque.
- Reservar extracción a microservicios para necesidades medidas. Los contratos ayudan, pero no convierten la extracción en un cambio automático: habrá que resolver datos, FK y transacciones distribuidas.

## Primer trabajo para empezar a programar

Completar E00–E03, luego F01–F04 del backlog. El primer resultado demostrable será una aplicación instalable con sesión, empresa/sucursal, migración V4 y contratos numéricos comprobados. Después construir el circuito recepción → apertura → venta/cobro → ticket → sincronización, con una prueba temprana de cada modalidad POS.

No esperar a terminar todas las pantallas administrativas para probar una venta offline. Ambas modalidades deben pasar la misma aceptación antes de declarar completo el MVP.

## Información pendiente y cuándo hace falta

| Información | Trabajo que condiciona | ¿Bloquea los planes? |
|---|---|---|
| Proyecto nuevo Ypy | Ruta y nombre confirmados | Resuelto |
| Sistemas operativos y navegadores del piloto | Instalador y matriz definitiva de soporte | No; usar Windows y navegador Chromium como objetivo provisional de pruebas, sin prometer otros entornos |
| Impresoras, lectores y balanzas reales | Integración y aceptación de periféricos | No; usar impresora simulada inicialmente, sin dar hardware por validado |
| Empresa y habilitación fiscal del piloto | Implementación concreta y aceptación fiscal | No para fundamentos; sí antes de emitir comprobantes reales |
| Cajas, artículos, ventas por minuto y duración habitual de desconexión | Carga, almacenamiento y dimensionamiento de bloques | No para fundamentos; sí para aprobar rendimiento y capacidad |
| Equipo de desarrollo y disponibilidad | Fechas y responsables nominales | No; este plan usa dependencias, no fechas inventadas |

## Alcance de la primera versión

Incluye catálogos, unidades/presentaciones, múltiples códigos de barras y lectura de peso/precio incorporado, precios básicos, compras directas y parciales, inventario/costos, conteos físicos y ajustes autorizados, POS con efectivo y medios mixtos registrados, caja con fondo inicial/ingresos/retiros/arqueo/cierre, cuentas pendientes, devoluciones, transferencias/despachos básicos, multimoneda, auditoría sensible, reportes esenciales y la modalidad fiscal habilitada elegida para el piloto. Registrar un pago con tarjeta no equivale a integrar un adquirente ni autorizarlo offline.

Contabilidad general, promociones complejas, fidelización, rutas avanzadas y procesos de otros rubros quedan fuera del primer lanzamiento. Los lotes/vencimientos se incluyen en la prueba comercial del supermercado; seriales se soportan en el contrato y se validan antes de habilitar artículos que los exijan.

## Revisión de recomendaciones incorporada

Se revisaron las sugerencias de la [conversación compartida](https://chatgpt.com/share/6aaa011e-6730-83e9-b48c-a1edbbf52456) y el usuario autorizó incorporar las convenientes. Se reforzaron contratos y aceptación sin cambiar el monolito modular, los tipos numéricos V4, la numeración autorizada ni la obligación de soportar ambas modalidades POS.

El backlog tiene 34 tareas y la matriz 30 escenarios. Se reutilizarán `gp_evento_entrada`, `gs_evento_salida`, `gs_auditoria`, `gp_terminal`, `gp_regla_balanza` y las tablas de caja/ajuste existentes; cualquier extensión se hará mediante migración durante implementación. Esta revisión solo modifica planes, no el SQL ni código de aplicación. Un piloto limitado a una modalidad no acredita aceptación de la otra.

