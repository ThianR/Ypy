# Plan de interfaz, formularios y personalización

Estado: propuesto, pendiente de implementación. Fecha: 2026-09-16.

## Objetivo

Conservar el estilo simple actual y establecer una experiencia común para formularios, tablas, filtros y búsquedas. Permitir que cada usuario personalice los colores y la presentación del sistema, con preferencias persistentes y recuperación del diseño predeterminado.

Este documento planifica trabajo; no acredita funcionalidades implementadas.

## Análisis de la situación actual

- La interfaz usa Vue 3, TypeScript, CSS y Dexie. `web/package.json` no incluye PrimeVue, aunque el plan inicial lo menciona.
- `web/src/styles.css` contiene variables de color y adaptación a 980 y 700 píxeles, junto con numerosos colores fijos.
- Catálogo y configuración implementan campos, errores y acciones directamente en sus vistas. Catálogo también mantiene su propio buscador con espera de 300 ms.
- `web/src/workspace.ts` concentra sesión, configuración y operaciones. Las preferencias visuales necesitan un estado propio para evitar ampliar esa concentración.
- El formulario de artículos permite altas y edición en demostración; su migración visual no habilita por sí sola el CRUD central pendiente.
- El backend ya declara GORM. Las preferencias persistentes seguirán la estructura de repositorios existente y las migraciones SQL del proyecto.

## Decisiones arquitectónicas

1. Mantener Vue y el estilo actual. Empezar con componentes pequeños, controles nativos y composición explícita. No incorporar una biblioteca completa como requisito de este plan. Si un control complejo exige una dependencia, justificarla y encapsularla detrás del componente común.
2. Cada módulo conserva sus reglas de negocio y llamadas de datos. La capa común controla presentación, interacción, estados y preferencias.
3. Compartir componentes y tipos; introducir definiciones de campos solo para patrones repetidos. Las pantallas complejas conservan componentes específicos.
4. Usar identificadores estables para pantallas, campos, secciones y columnas. Las preferencias se vinculan a esos identificadores, nunca al texto visible ni a posiciones del DOM.
5. Guardar preferencias declarativas y versionadas. No admitir CSS, HTML o JavaScript arbitrario como personalización.
6. La personalización modifica presentación. Los permisos y la obligatoriedad de negocio se verifican también en el backend.

## Estándar de interacción

| Elemento | Comportamiento compartido |
|---|---|
| Formulario | Título, descripción breve, secciones, estados y acciones consistentes |
| Campo | Etiqueta visible, ayuda, valor, indicación de obligatorio y error asociado |
| Validación | Validar al salir de un campo editado y al enviar; enfocar el primer error y mostrar resumen cuando corresponda |
| Obligatorio | Texto visible y semántica accesible; distinguir vacío de valores válidos como cero o falso |
| Guardado | Mostrar progreso, bloquear envíos simultáneos y conservar valores ante error |
| Cambios pendientes | Detectar edición y advertir al cerrar, navegar o cambiar de registro |
| Búsqueda | Limpiar, estado de búsqueda, resultado vacío y control de respuestas fuera de orden |
| Filtros | Operadores según el tipo, aplicar, limpiar y resumen de filtros activos |
| Tabla | Orden, paginación, columnas, estados y acciones de fila uniformes |
| Mensajes | Español profesional, explicación concreta y acción de recuperación |

La operación determina las acciones disponibles: no todos los formularios tendrán borrador o eliminación. La confirmación comercial del POS seguirá su flujo específico. El bloqueo visual de doble envío no sustituye la idempotencia de las operaciones del backend.

## Componentes y organización propuesta

- `web/src/ui/`: botones, campos, secciones, contenedor de formulario, acciones, diálogos, avisos, tablas, filtros y buscadores.
- `web/src/composables/`: estado de formulario, validación, cambios pendientes y consultas de tablas.
- `web/src/preferences/`: tipos, resolución de preferencias, validación, almacenamiento y migraciones de versión.
- `web/src/styles/`: variables semánticas, tema predeterminado y reglas adaptables comunes.
- `web/src/components/`: vistas de negocio que consumen los componentes anteriores.

Crear archivos y componentes según se migren usos reales. Evitar un componente único con todas las variantes o un registro global de extensiones antes de necesitarlo.

### Tipos de campo

Primera entrega: texto, texto largo, contraseña, entero, decimal, importe, selección, selección múltiple, casilla, fecha, hora y fecha con hora. Segunda entrega: búsqueda de entidades y rangos reutilizables en filtros.

Cada campo tendrá identificador, etiqueta, ayuda, obligatoriedad, estado de lectura, restricciones de valor y mensajes de error. Los campos compuestos admitirán contenido específico mediante composición de Vue.

- Fecha sin hora: transportar `YYYY-MM-DD`; no convertirla a UTC ni desplazarla por zona horaria.
- Hora sin fecha: transportar `HH:mm` o `HH:mm:ss`, según el contrato del campo.
- Fecha con hora: declarar si representa un instante o una hora local de negocio. Para instantes, intercambiar una representación con zona o desplazamiento explícito y mostrarla en la zona configurada. Para horarios locales, conservar fecha, hora y zona cuando corresponda.
- Formato visible: español de Paraguay por defecto. Separar formato de visualización y valor enviado a la API.
- Importes: reutilizar la aritmética decimal exacta existente; no transformar los importes del negocio en cálculos binarios de punto flotante.

### Tablas, filtros y buscadores

- Columnas con identificador, tipo, etiqueta, alineación y capacidades de orden y filtro.
- Orden, visibilidad, ancho y fijación de columnas; densidad cómoda o compacta y tamaño de página.
- Filtros de texto, selección, estado, intervalo numérico y rango de fecha o fecha con hora.
- Primera versión: combinación de filtros mediante AND. Expresiones avanzadas se incorporarán únicamente cuando exista un caso funcional concreto.
- Reiniciar paginación al modificar filtros; manejar límites y resultados vacíos después de eliminar o modificar registros.
- Dos modos explícitos: consulta local para datos completos de demostración y consulta remota para datos paginados. No presentar un filtro sobre una página cargada como búsqueda global del servidor.
- Contrato remoto con búsqueda, filtros tipados, orden y paginación. El backend valida campos y operadores permitidos y aplica el ámbito y los permisos antes de devolver resultados.
- Cancelar o descartar respuestas antiguas en búsquedas remotas; ofrecer reintento ante error.
- Guardar vistas con nombre, filtros y columnas. Separar esas preferencias de la selección temporal de filas.
- En móvil, ofrecer resumen por registro con acceso al detalle. Mantener desplazamiento horizontal contenido cuando la comparación entre columnas sea necesaria.

## Colores y vistas personalizables

### Paleta predeterminada

Conservar como punto de partida la identidad actual:

| Función | Valor inicial |
|---|---|
| Color principal | `#154D47` |
| Principal al interactuar | `#0D3733` |
| Fondo general | `#F5F6F2` |
| Superficie | `#FFFFFF` |
| Texto principal | `#1D2A2A` |
| Texto secundario | `#687474` |
| Borde | `#DDE3DE` |
| Acento | `#D87B45` |

Estos valores son una base visual, no una certificación de contraste. Definir pares de texto y fondo y comprobarlos durante la implementación.

Reemplazar todos los colores fijos por variables semánticas. El editor permitirá configurar fondos, superficies, textos, bordes, botones, menú, cabeceras y filas de tablas, campos, foco, selección y estados de éxito, advertencia, información y error. Incluir colores de texto sobre cada superficie y estados de interacción.

Ofrecer edición simple por paleta y edición avanzada de cada variable, con vista previa, aplicar, cancelar y restaurar. Conservar texto y otras señales además del color para expresar estados. Detectar combinaciones ilegibles y ofrecer corrección antes de aplicarlas.

### Personalización de vistas

- Sistema: tema, densidad, escala tipográfica y preferencia de menú.
- Formulario: orden de secciones y campos, distribución dentro de una cuadrícula adaptable, ancho permitido, secciones plegadas y campos opcionales visibles.
- Tabla: columnas visibles, orden, ancho, fijación, tamaño de página, filtros y ordenamiento inicial.
- Pantalla: vistas guardadas con nombre, duplicación, cambio de nombre, eliminación y selección de vista inicial.
- Colores particulares por pantalla mediante sustituciones de las mismas variables semánticas.

El editor de vista tendrá controles de subir, bajar y mover a sección; el arrastre será complementario para permitir uso con teclado y móvil. La previsualización mostrará escritorio y móvil antes de guardar.

Un campo obligatorio editable no podrá ocultarse si necesita intervención del usuario. Solo podrá quedar oculto cuando el sistema resuelva válidamente su valor. Los campos sin permiso nunca aparecerán por restaurar una vista. Los grupos con dependencias podrán limitar su reordenamiento.

Personalización completa significa cubrir los colores y opciones de presentación de los componentes del sistema. Cambiar esquemas de datos, agregar reglas ejecutables o crear campos de negocio nuevos requiere otro alcance.

### Prioridad y persistencia

Orden de aplicación, de menor a mayor prioridad:

1. Valores predeterminados del producto.
2. Valores predeterminados de la empresa, administrados con permiso específico.
3. Preferencias generales del usuario dentro de esa empresa.
4. Preferencias de la vista seleccionada por ese usuario.

Aplicar finalmente permisos, restricciones de campos y adaptación al dispositivo. Al cambiar el tema general, las vistas heredan todo lo que no tengan personalizado explícitamente. Permitir restablecer una propiedad, una vista o todas las preferencias personales.

Persistencia propuesta: configuración versionada en PostgreSQL, mediante repositorios GORM y migraciones SQL explícitas. Preferencias personales identificadas por empresa, usuario y pantalla o ámbito global; valores de empresa almacenados por separado. Elegir tablas existentes compatibles antes de crear nuevas. Usar una revisión para detectar ediciones concurrentes y evitar sobrescrituras silenciosas entre dispositivos.

El servidor obtiene empresa y usuario de la sesión y valida propiedad, permisos, identificadores, tamaños y valores. Publicar el contrato de consulta, guardado y restablecimiento en OpenAPI antes de conectar el editor.

Mantener una caché local aislada por empresa y usuario. En demostración usar un espacio separado. Si no hay conexión, aplicar localmente y marcar cambios pendientes; al reconectar sincronizar preferencias sin mezclarlas con la cola de ventas. Ante conflicto ofrecer conservar la versión local o la del servidor.

Las preferencias guardan disposición y filtros elegidos expresamente, no contraseñas, tokens ni valores editados de los formularios. Al cerrar sesión, retirar el estado activo para evitar que el siguiente usuario herede preferencias personales. Ante una versión incompatible, recuperar los valores predeterminados e informar sin impedir el acceso.

## Implementación por entregas

| Entrega | Dependencia | Trabajo | Criterio de aceptación |
|---|---|---|---|
| UI01. Inventario y contrato visual | Ninguna | Inventariar pantallas, campos, colores y estados; asignar identificadores estables; documentar decisiones sobre componentes | Catálogo de elementos y ejemplo de formulario, tabla y vista móvil revisables |
| UI02. Tema y estructura adaptable | UI01 | Extraer variables semánticas, paleta predeterminada, menú y contenedores; eliminar colores fijos de las vistas migradas | Cambiar el tema afecta a todos los componentes migrados; el estilo predeterminado conserva su identidad |
| UI03. Formularios comunes | UI02 | Campos tipados, validación, foco, errores, cambios pendientes y acciones; migrar configuración y edición de artículo como pilotos | Los dos formularios comparten comportamiento y permiten verificar fechas y obligatorios en una página interna de ejemplos |
| UI04. Tablas y consultas | UI03 | Tabla, buscador, filtros, paginación y modos local/remoto; migrar catálogo e historial | Filtros y orden actúan sobre el conjunto correcto; respuestas antiguas no reemplazan búsquedas nuevas |
| UI05. Preferencias persistentes | UI01, UI04 | Modelo versionado, migración, repositorio GORM, contrato OpenAPI, API, caché y resolución de prioridades | Dos usuarios y empresas mantienen preferencias aisladas; recarga, reconexión y conflictos tienen comportamiento verificable |
| UI06. Editor de apariencia | UI02, UI05 | Paleta simple y avanzada, vista previa, validación, aplicar, cancelar y restablecer | Tema personal persiste entre sesiones y dispositivos; es posible recuperar la apariencia inicial |
| UI07. Editor de vistas | UI03, UI04, UI05 | Orden y distribución de campos, secciones, columnas, filtros y vistas con nombre | Cambios persisten por pantalla; obligatorios y permisos conservan su funcionamiento |
| UI08. Migración y cierre | UI06, UI07 | Migrar acceso, venta y pantallas restantes; documentar patrón, personalización y extensión global | Pantallas actuales usan el estándar donde corresponde y pasan la matriz de aceptación |

Cada entrega incluye comprobación de teclado, adaptación y estados. UI08 verifica la integración completa. No asignar fechas sin medir el primer piloto y conocer la capacidad disponible.

## Validación y definición de terminado

- Revisar anchos de 320, 375, 768, 1024 y 1440 píxeles, orientación horizontal y zoom al 200 %. Sin recortes de acciones ni desplazamiento global horizontal.
- Comprobar formularios y diálogos con teclado, etiquetas asociadas, foco visible, errores anunciados y teclado virtual móvil. Las acciones fijas no deben tapar campos ni mensajes.
- Probar fecha sin hora sin cambios de día, año bisiesto, rangos invertidos, valores vacíos, cero, decimales y fecha con hora en distintas zonas.
- Probar filtro remoto, paginación y respuestas fuera de orden; distinguir error, búsqueda en curso y ausencia de resultados.
- Probar cancelación del editor, restauración, herencia de tema, vista corrupta y evolución de identificadores después de una actualización.
- Probar aislamiento entre usuarios y empresas, permisos de valores empresariales y edición concurrente entre dispositivos.
- Probar guardado de preferencias sin conexión y posterior reconciliación, sin efectos sobre operaciones comerciales.
- Ejecutar tipado, compilación y pruebas existentes; agregar pruebas de comportamiento para los casos anteriores y comprobaciones de integración para API y persistencia.
- Documentar un ejemplo real de nuevo formulario y una mejora global, verificando que no sea necesario editar cada pantalla para incorporarla.

El primer resultado revisable será configuración y edición de artículos usando campos comunes, junto con catálogo e historial compartiendo tabla y filtros. La personalización completa quedará aceptada al finalizar UI08, no al entregar únicamente el selector de colores.
