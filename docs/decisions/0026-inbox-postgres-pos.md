# Decisión 0026: inbox PostgreSQL para operaciones POS

## Decisión

Se agrega `internal/poscentral.PostgresStore` como adaptador de persistencia de la
entrada POS sobre `erp_v4.gp_evento_entrada`. La operación se identifica por UUID,
se valida el JSON y se guarda su SHA-256 en `contenido_hash`; los reintentos con el
mismo UUID y contenido son idempotentes y un contenido distinto devuelve conflicto.

## Alcance

Este adaptador acepta la operación en el inbox. La aplicación comercial (venta,
stock, numeración, cobro y auditoría) debe ejecutarse en una transacción posterior,
con sus propias reglas y permisos; no se simula esa confirmación dentro del inbox.

## Verificación

Compilación y `scripts/verify-all.ps1` pasan después de incorporar el adaptador.
