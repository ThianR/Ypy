# ADR 0009 — Contexto de auditoría

Se reutiliza `gs_auditoria` y se agregan UUID de operación, terminal, instalación, motivo, request y correlación. Los secretos no se guardan en antes/después. La aplicación debe escribir auditoría y cambio sensible en la misma transacción; intentos rechazados se registran por el canal de seguridad correspondiente.

