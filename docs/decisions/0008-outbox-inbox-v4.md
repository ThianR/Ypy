# ADR 0008 — Idempotencia de eventos V4

Se reutilizan `gp_evento_entrada` y `gs_evento_salida`. La migración agrega hash canónico, resultado, fecha de procesamiento y reintentos al inbox, además de error y reintentos al outbox. El UUID y las restricciones existentes siguen siendo la identidad durable; no se crea un segundo inbox/outbox.

