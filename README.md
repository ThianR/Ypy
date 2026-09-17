# Ypy

## Inicio rápido y URL de acceso

Requisitos: Windows, PowerShell, Node.js, npm y PostgreSQL para persistencia central. La persistencia Go utiliza GORM; el agente local usa el controlador SQLite puro en Go. El runtime portable de Go está incluido en `.tools/go`.

### Backend

Desde una terminal PowerShell ubicada en `Ypy`:

```powershell
$env:GOROOT = $null
$env:Path = ".tools\go\bin;$env:Path"
go run .\cmd\erp
```

El backend escucha en `http://127.0.0.1:8080`. Para PostgreSQL, configure antes `YPY_DATABASE_URL`.

### Frontend

En una segunda terminal:

```powershell
Set-Location .\web
npm install
npm run dev
```

La interfaz queda normalmente en `http://localhost:5173`; Vite muestra la URL exacta al iniciar.

| URL | Uso |
|---|---|
| `http://localhost:5173` | Interfaz web durante desarrollo |
| `http://127.0.0.1:8080/health` | Comprobar proceso activo |
| `http://127.0.0.1:8080/ready` | Comprobar servicio listo |
| `http://127.0.0.1:8080/api/v1/auth/login` | Inicio de sesión de la API |
| `http://127.0.0.1:8080/api/v1/catalog/items` | Búsqueda de artículos y precios vigentes |

### Uso de la interfaz

La pantalla de acceso ofrece dos recorridos:

- Acceso central: utiliza el usuario de la empresa y valida la sesión contra PostgreSQL.
- Exploración de demostración: abre un espacio local con artículos, existencias y ventas de práctica guardados en el navegador.

En el entorno de laboratorio incluido, el acceso central de demostración es `demo` con contraseña `demo123`, empresa `1`, sucursal `1` y terminal `1`. Esas credenciales son únicamente para laboratorio y no deben utilizarse en producción.

Después de entrar, el menú permite acceder a Inicio, Punto de venta, Catálogo de artículos, Historial de ventas, Configuración y Guía de uso. El circuito completo de demostración es: seleccionar un artículo, revisar el carrito, confirmar el medio de pago y consultar el comprobante en el historial.

La demostración no emite comprobantes fiscales ni envía ventas al servidor. La conexión central ya permite iniciar y cerrar sesión y consultar precios; la venta central, la edición central de artículos, caja, compras e inventario administrativo requieren completar sus pantallas e integración.

### Agente POS local

El agente usa SQLite y no publica una URL remota:

```powershell
$env:YPY_POS_DB = ".\data\pos-terminal-01.db"
$env:YPY_POS_CENTRAL_URL = "http://127.0.0.1:8080"
$env:YPY_POS_EMPRESA_ID = "1"
$env:YPY_POS_TERMINAL_ID = "2"
go run .\cmd\pos-agent
```

### Pruebas

```powershell
go test ./...
go build ./cmd/...
powershell -File .\scripts\verify-all.ps1
powershell -File .\scripts\smoke-http.ps1
```

## Estado

Implementación incremental en curso. El esqueleto, contratos, frontend POS, agente SQLite, adaptadores centralizados y migraciones base ya están versionados; las etapas restantes se cierran progresivamente con pruebas y verificación integral.

## Procedencia de la base

La estructura inicial procede de [creacionV4.sql](../creacionV4.sql) y su [informe de diseño](../V4_DISENO_Y_REVISION.md). El nombre histórico Genesis en esos archivos identifica su procedencia. La denominación del producto es Ypy; no se renombra automáticamente el esquema SQL `erp_v4` ni sus identificadores.

## Comandos históricos de verificación

Con el runtime portable incluido en `.tools/go`:

```powershell
$env:GOROOT=$null
$env:Path=".tools\go\bin;$env:Path"
go test ./...
go build ./cmd/...
powershell -File scripts\verify-all.ps1
powershell -File scripts\smoke-http.ps1
powershell -File scripts\bench.ps1
```

Frontend: `cd web; npm install; npm run dev`. El ERP escucha en `:8080`; el agente local usa `YPY_POS_DB` para seleccionar su archivo SQLite. `YPY_POS_BACKUP` crea una copia SQLite verificada y no sobrescribe destinos existentes. `YPY_DATABASE_URL` habilita persistencia central PostgreSQL y la integración comercial POS. `YPY_POS_CENTRAL_URL`, `YPY_POS_EMPRESA_ID` y `YPY_POS_TERMINAL_ID` habilitan la sincronización continua del agente; `YPY_POS_SYNC_INTERVAL_SECONDS` configura el intervalo (1–3600 s, por defecto 30).

El POS consulta el catálogo central mediante `GET /api/v1/catalog/items?empresa_id=<id>&lista_id=<id>&as_of=<RFC3339>&q=<texto>`. Cuando PostgreSQL o la red no están disponibles, conserva el catálogo local de demostración y permite continuar registrando ventas offline. El bootstrap de balanzas consulta `GET /api/v1/catalog/scale-rules?empresa_id=<id>` y guarda las reglas en el caché Dexie del POS web.

Para aplicar migraciones de forma reproducible, define `YPY_DATABASE_URL` y ejecuta `.\scripts\apply-migrations.ps1`. Luego puedes ejecutar `.\scripts\verify-db-contract.ps1` para comprobar, en modo solo lectura, las tablas y función comercial críticas. Las migraciones V4 deben aplicarse únicamente sobre una base de laboratorio o una instalación administrada.

Para generar una entrega sin secretos ni bases locales: `.\scripts\package-release.ps1`. El ZIP incluye `release-manifest.json` con la versión indicada por `YPY_RELEASE_VERSION` (o `development`) y SHA-256 de ambos binarios.
