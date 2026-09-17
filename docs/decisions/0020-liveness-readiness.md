# ADR 0020 — Estado operativo

`/health` indica que el proceso está vivo y `/ready` que el despliegue puede recibir trabajo. Ninguno expone credenciales, configuración sensible o datos de empresas; las comprobaciones de dependencias se incorporarán a readiness al conectar PostgreSQL.

