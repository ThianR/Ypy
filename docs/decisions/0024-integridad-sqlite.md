# ADR 0024 — Integridad de SQLite

El agente verifica `PRAGMA integrity_check` antes de usar una base local recuperada. Una base que no responde `ok` no se presenta como operativa; primero se preserva para diagnóstico y se aplica el procedimiento de recuperación.

