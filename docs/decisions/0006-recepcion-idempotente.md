# ADR 0006 — Recepción idempotente

La recepción se identifica por su documento y se aplica una sola vez. El saldo se consulta desde el estado actual del libro, no desde un saldo inicial congelado. La implementación inicial es un núcleo de prueba; la autoridad definitiva seguirá siendo V4 al integrar V01.

