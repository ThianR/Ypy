# ADR 0022 — Contrato HTTP versionado

El endpoint POS se documenta bajo `/api/v1` con operación explícita, `X-Request-ID` y respuestas 400/409/500 diferenciadas. Los identificadores BIGINT viajan como cadenas y el contrato de integración no expone filas SQL.

