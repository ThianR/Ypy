# ADR 0019 — Verificación local reproducible

`scripts/verify-all.ps1` es la puerta local inicial: valida formato, pruebas, análisis `go vet` y compilación Go, límites de módulos, typecheck/tests/build del frontend, hash del baseline y fixture. Las pruebas PostgreSQL de migraciones siguen requiriendo la instancia aislada de laboratorio y no se ejecutan contra producción.
