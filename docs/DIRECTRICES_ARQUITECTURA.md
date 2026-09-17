# Directrices de Estructura de Proyecto y Arquitectura

Este documento establece las reglas y estándares de organización para el repositorio Ypy, asegurando una base de código limpia, testeable y sostenible.

## 1. Organización del Directorio Raíz

La estructura principal del proyecto se rige por las siguientes responsabilidades:

- **`cmd/`**: Contiene únicamente los puntos de entrada (ejecutables) de la aplicación (`erp`, `pos-agent`). **Prohibido** incluir lógica de negocio, montajes complejos de rutas HTTP o acceso directo a bases de datos en los archivos de este directorio.
- **`internal/`**: Contiene el código privado y de negocio de la aplicación. Se subdivide en:
  - **`bootstrap/`**: Responsable del ensamblado, inyección de dependencias, configuración de middlewares y montaje del servidor HTTP.
  - **`platform/`**: Infraestructura compartida, utilidades técnicas y código agnóstico al negocio (ej. conectores de base de datos, clientes HTTP base, utilería para eventos).
  - **`modules/`**: Contiene los distintos dominios de negocio (Contextos Delimitados). Cada módulo debe encapsular su dominio, casos de uso y adaptadores.
  - **`poslocal/`**: Lógica de persistencia (SQLite), periféricos (impresión) y sincronización del agente POS que corre en las sucursales.
  - **`poscentral/`**: Lógica de ingesta centralizada y validación de operaciones que provienen del POS.
- **`contracts/`**: Documentación y especificaciones (JSON/HTTP/SQL) estables y compartidas.
- **`web/`**: Proyecto frontend independiente (Vue 3). Este directorio no debe contener dependencias o subdirectorios propios del backend.
- **`docs/`**: Documentación técnica del proyecto, esquemas y notas de diseño.

## 2. Reglas de Implementación en Módulos (`internal/modules/`)

Cada módulo debe adherirse a los siguientes principios de encapsulamiento:

- **Interfaces Claras**: El subpaquete `api` (o `contracts`) de cada módulo debe exponer únicamente las interfaces, tipos de transferencia (DTOs) y las definiciones públicas. 
- **Separación de Persistencia**: **Evitar** colocar implementaciones concretas de PostgreSQL (ej. `gorm.DB`) y consultas crudas dentro de los mismos archivos que definen la lógica pública o contratos de transporte. Se sugiere colocar los adaptadores de persistencia en subpaquetes como `postgres` o `store`.
- **Independencia Funcional**: Los consumidores externos a un módulo no deben importar ni acoplarse a las entidades internas de persistencia de otro módulo.

## 3. Puntos de Entrada Livianos (`cmd/`)

El contenido de un archivo en `cmd/` (`main.go`) debe limitarse a:

- Leer variables de entorno y configuración.
- Inicializar conexiones básicas y logs.
- Invocar los constructores o funciones `Run()` correspondientes en los paquetes internos (ej. `bootstrap.NewServer(...)`).
- Escuchar señales de interrupción del sistema operativo para permitir un apagado ordenado (Graceful Shutdown).

## 4. Políticas de Limpieza e Higiene del Repositorio

- **Archivos Temporales y Artifacts**: No dejar archivos empaquetados (`.zip`), volcado de logs de depuración (`.log`, `.xml` crudo) en la raíz del proyecto. Para depuración, utilizar una carpeta temporal añadida al `.gitignore` o escribir en las rutas correspondientes del OS.
- **Contexto de Comandos**: Ejecutar comandos y scripts (ej. migraciones o pruebas) desde el contexto y directorio adecuado. Evitar generar directorios "fantasmas" (ej. en `web/`).
