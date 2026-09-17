# ADR 0021 — Cola de impresión

La impresión es posterior a la persistencia comercial y tiene cola propia. Sin respuesta del periférico, el ticket queda `UNKNOWN` y puede reintentarse; nunca se crea otra venta ni se revierte la operación por una falla de impresión.

