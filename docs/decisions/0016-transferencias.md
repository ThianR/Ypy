# ADR 0016 — Transferencias de stock

Una transferencia identifica origen, destino, artículo y cantidad; su ID es idempotente. El origen debe tener saldo suficiente y el despacho posterior de una venta no debe volver a descontar stock.

