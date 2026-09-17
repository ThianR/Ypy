# ADR 0004 — PostgreSQL de laboratorio

Las migraciones se validaron el 16-09-2026 en una instancia PostgreSQL 17.11 limpia, inicializada con el runtime portátil, puerto 55435 y datos dentro de `.lab-pgdata`. Se aplicaron `0001_baseline_v4.sql`, `0002_identidad.sql` y `0003_eventos_idempotencia.sql` con `ON_ERROR_STOP=1`; se verificaron `erp_v4.ypy_usuario_sesion` y las columnas de idempotencia del inbox. Las migraciones fijan `search_path` explícitamente. La instancia se detuvo al finalizar. `.lab-pgdata` es solo laboratorio y no contiene datos de producción.
