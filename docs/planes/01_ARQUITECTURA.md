# Arquitectura del monolito modular

## 1. Organización de ejecución

Backend central Go con API HTTP y trabajadores internos para outbox, sincronización fiscal y reportes. Un despliegue central y PostgreSQL compartido. Los trabajadores tendrán concurrencia acotada y apagado ordenado; una caída del generador PDF no debe interrumpir una transacción comercial.

Frontend Vue/PrimeVue común: administración y POS. En desarrollo, Vite y Air. En distribución, frontend compilado con Go Embed. Agente local Go opcional por terminal, con SQLite y adaptador de impresión. La alternativa solo navegador usa Dexie/PWA. No se requiere un bus externo ni bases por módulo para comenzar.

La persistencia Go utiliza GORM con los controladores de PostgreSQL y SQLite puro en Go. Las migraciones SQL y las funciones V4 mantienen la autoridad sobre el esquema y las reglas transaccionales; GORM no ejecuta `AutoMigrate` sobre la base existente. Las consultas complejas y las funciones V4 se encapsulan mediante `Raw` y `Exec`. Versionar migraciones, contratos y dependencias.

## 2. Propiedad de módulos

| Módulo | Propiedad y responsabilidad | Contratos principales |
|---|---|---|
| identidad | Usuarios, empresas/sucursales, roles y sesión | Autorizar operación y contexto de empresa/terminal |
| maestros | Terceros, monedas, cotizaciones e impuestos | Consultar datos/versiones; convertir importes |
| catalogo | Artículos, presentaciones, barras y precios | Resolver artículo y precio comercial |
| numeracion | Tipos, talonarios, habilitaciones, números y bloques | Reservar bloque, consumir número y conservar anulación |
| inventario | Saldos, reservas, kardex, valoración y ajustes | Confirmar efecto de stock y consultar disponibilidad |
| compras | Órdenes, recepciones, facturas y aplicaciones | Recepción, compra directa y conciliación de factura |
| ventas | Documentos de venta, devoluciones y notas | Confirmar venta, devolver y anular |
| tesoreria | Caja, cobros/pagos, cuotas y aplicación de créditos | Cobrar, pagar y cerrar caja |
| logistica | Preparación, despacho, entrega y transferencias coordinadas | Preparar/despachar sin repetir el efecto de stock |
| pos | Terminales, permisos offline, eventos y conflictos | Descargar configuración e integrar operación local |
| fiscal | Adaptadores, documentos enviados, respuestas y eventos fiscales | Preparar, enviar y consultar estado según modalidad |
| reportes | Consultas y exportaciones; no dueño de documentos | Reportes por empresa con trazabilidad |

Los prefijos SQL no bastan para asignar propiedad: por ejemplo, catálogo y saldos pueden compartir prefijo `gi_`. En E01 producir un manifiesto con cada tabla y función V4, propietario y dependencias. Cada objeto tendrá un único responsable. Los coordinadores de procesos no se convierten en propietarios de todas las tablas que utilizan.

## 3. Reglas de dependencia comprobables

- Módulos exponen contratos pequeños en `api`; los consumidores importan esos contratos, nunca repositorios ni entidades internas del vecino.
- Interfaces usan identificadores y datos de negocio, no filas SQL, conexiones, tipos HTTP ni detalles de Vue.
- Una capa `workflows` coordina procesos compuestos. Los módulos no dependen de esa capa ni forman ciclos entre sí.
- Infraestructura implementa persistencia, impresión y transporte. `platform` contiene herramientas técnicas; no acumula reglas de compras/ventas.
- Evitar escribir tablas ajenas desde Go. Las funciones V4 que ya hacen operaciones cruzadas son excepciones existentes, inventariadas y encapsuladas por adaptadores; no ampliar estas excepciones sin una decisión registrada.
- Consultas cruzadas necesarias para reportes viven en adaptadores de lectura identificados. No reutilizarlas como vía de modificación.
- Verificar importaciones prohibidas y ciclos en CI. El manifiesto SQL y la revisión de migraciones cubren dependencias que el compilador Go no detecta.

## 4. Transacciones y eventos

Confirmar venta contado exige una transacción PostgreSQL que registre venta, número, stock/costo, cobro y evento durable. El coordinador abre la unidad de trabajo; adaptadores usan la misma transacción sin exponer `pgx.Tx` en los contratos de negocio. Un error de cobro revierte el conjunto.

En V4 `gp_integrar_venta` no completa por sí sola todo el circuito de caja. P03 debe revisar su manejo de errores/transacciones y construir una integración atómica venta+cobro+acuse. No marcar una venta como integrada mientras el cobro está pendiente accidentalmente. El estado de conflicto durable se registra sin efectos comerciales parciales.

Completar el outbox PostgreSQL existente (`gs_evento_salida`) y la recepción durable (`gp_evento_entrada`): evento y cambio de negocio se guardan juntos. El trabajador publica/procesa con reintentos; los consumidores deduplican por evento. La entrega puede repetirse. No utilizar eventos asíncronos para separar hoy el cobro y el descuento de stock que deben confirmar juntos. Ver [contrato de sincronización](05_SINCRONIZACION_POS.md).

Eventos propuestos: VentaConfirmada, RecepcionConfirmada, CobroRegistrado, ExistenciaActualizada, DocumentoFiscalPendiente. Incluir UUID, empresa, versión, origen, correlación y fecha. Reportes/fiscal pueden consumirlos tras commit. Las llamadas externas no permanecen dentro de la transacción SQL; cada modalidad fiscal define estados y cuándo se puede entregar el comprobante al cliente.

## 5. Estructura propuesta de código

Raíz confirmada: `D:\Personales\SISTEMAS\Genesis\Ypy`. Proyecto nuevo, sin reutilización implícita de Venture. Las rutas siguientes son relativas a Ypy y se crearán durante la implementación.

```text
cmd/erp/                     arranque central
cmd/pos-agent/               arranque del agente local
internal/bootstrap/         composición de dependencias
internal/platform/          HTTP, SQL, logs, reloj, criptografía
internal/workflows/         procesos transaccionales compuestos
internal/modules/<modulo>/api/
internal/modules/<modulo>/internal/domain/
internal/modules/<modulo>/internal/application/
internal/modules/<modulo>/internal/adapters/
internal/poslocal/           persistencia SQLite y periféricos
web/src/modules/            pantallas por módulo
web/src/pos/                 flujo común de caja
web/src/pos/adapters/        Dexie y cliente del agente local
contracts/http/             OpenAPI y errores
contracts/events/           eventos versionados
contracts/fixtures/         casos numéricos y comerciales compartidos
db/migrations/              baseline V4 y evoluciones
tests/integration/          PostgreSQL y procesos
tests/e2e/                  POS navegador y agente
deploy/                     servicio, configuración y empaquetado
docs/decisions/             decisiones y excepciones
```

## 6. Base V4 y cambios incrementales

Conservar V4 como baseline reproducible, con hash y procedencia. E02 debe verificar si el ejecutor de migraciones permite el BEGIN/COMMIT y `search_path` del script; no anidar ciegamente transacciones. Separar catálogos globales de configuración real de una empresa. No ejecutar datos de pruebas en producción.

Las nuevas capacidades (outbox, sesiones, integración de cobros offline, cotizaciones fijadas por operación si faltan) se agregan mediante migraciones posteriores. No editar una migración ya aplicada. Antes de migrar datos antiguos hace falta un proyecto específico; la V4 actual instala una base nueva.

## 7. Extracción futura a servicios

Primero medir cuellos de botella y necesidades de despliegue. Reportes, procesamiento fiscal o integraciones externas suelen ser candidatos de menor acoplamiento, pero la elección depende de mediciones.

Para extraer un módulo: estabilizar contrato, inventariar tablas/FK/funciones cruzadas, crear proyecciones de lectura, trasladar propiedad de datos, sustituir llamadas locales por transporte y resolver fallos parciales. Cuando ya no exista transacción compartida, diseñar compensaciones y estados explícitos; comprobar conciliación, reintentos y recuperación antes del corte.

El bloqueo por empresa de V4 y sus funciones cruzadas constituyen deuda conocida para escalado/extracción. No quitar FK ni consistencia actual solo para simular independencia. El objetivo del MVP es reducir acoplamiento nuevo y dejar documentado el existente.

## 8. Invariantes del sistema

1. Una operación aplicada no repite venta, cobro, número ni stock al retransmitirse.
2. Cada documento pertenece a una empresa; toda referencia y permiso respeta ese ámbito.
3. Una operación confirmada conserva sus cantidades, precios, impuestos, moneda, cambio y evidencia originales; correcciones generan operaciones relacionadas, no eliminación física.
4. Todo movimiento de inventario tiene origen trazable. Central decide el kardex y valoración definitivos; ajustes de costo posteriores se registran por separado, sin reescribir movimientos confirmados.
5. Todo movimiento de caja física pertenece a una sesión; no significa que todos los pagos bancarios del ERP deban pertenecer a caja POS.
6. Terminal, instalación y sesión de caja son identidades distintas. Una sesión activa no cambia de empresa/sucursal por editar configuración.
7. Una factura de compra vinculada a recepción y un despacho vinculado a venta no duplican el movimiento de stock.
8. Un estado fiscal pendiente no se presenta como aceptado. Agotamiento de rangos no habilita numeración fiscal inventada.
9. Las correcciones sensibles requieren permiso, motivo y auditoría; un flag de funcionalidad no concede permisos.
10. Sustituir una función SQL por Go requiere pruebas de caracterización, equivalencia y motivo concreto; no se hace por preferencia estética.

Caja permanece como subdominio explícito de tesorería, con contratos propios. Auditoría es una capacidad transversal con propietario identificado; no se crean nuevos módulos o capas solo por cada recomendación recibida.

