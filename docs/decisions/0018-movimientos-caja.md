# ADR 0018 — Movimientos manuales de caja

Ingresos y retiros requieren sesión, tipo, importe positivo y motivo. Se registran por identidad idempotente y nunca generan automáticamente ventas, deudas o pagos a proveedores.

