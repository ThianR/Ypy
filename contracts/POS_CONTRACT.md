# Contrato POS común (E03)

Los adaptadores Dexie/PWA y Go/SQLite implementan el mismo sobre: `operationId`, empresa, sucursal, terminal, sesión, secuencia, versión, fecha comercial, moneda, total y pagos. Los identificadores y valores financieros se transportan como cadenas; no se usa `Number`/`float64` para aritmética.

La transacción local guarda operación, consumo de número, cobro, stock provisional y cola. La impresión ocurre después y su respuesta incierta no revierte la operación. Los reintentos con el mismo UUID y contenido devuelven el resultado previo; el mismo UUID con contenido distinto pasa a conflicto.

