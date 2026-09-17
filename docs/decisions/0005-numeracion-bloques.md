# ADR 0005 — Numeración por bloques

La numeración POS se reserva por terminal en bloques exclusivos. El asignador no calcula el siguiente número a partir del máximo documental; un terminal extranjero y un bloque agotado son errores explícitos. La persistencia, firma de permisos, reserva y conciliación central se implementan sobre las tablas V4 en F04/P03.

