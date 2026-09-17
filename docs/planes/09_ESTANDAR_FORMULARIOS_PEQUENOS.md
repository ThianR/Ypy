# Estándar de formularios pequeños

Piloto: Unidades de medida. Aplica a catálogos con pocos campos, sin detalles anidados.

Este patrón es una variante del [estándar común de formularios](10_ESTANDAR_COMUN_FORMULARIOS.md).

## Decisión de diseño

La consulta ocupa el ancho disponible. Crear y editar abre un panel modal lateral,
independiente del tamaño de la tabla. En móvil ocupa el ancho completo.
El cuerpo del panel se desplaza; título y acciones permanecen visibles.

Se conserva la identidad de Ypy: verde #154D47, fondo #F5F6F2, superficie #FFFFFF,
texto #1D2A2A y borde #DDE3DE, mediante las variables del tema. Tipografía existente,
etiquetas alineadas a la izquierda, ayudas de lectura cómoda, sin rótulos decorativos.

    Título del catálogo                                      [Nuevo]
    Ayuda (mostrar/ocultar)
    Búsqueda | Filtros específicos
    Tabla a todo el ancho | Acciones fijadas a la derecha
    Filas por página | Rango / total | Anterior / Siguiente

    Al crear o editar: panel lateral
    Título | Cerrar
    Campos y ayuda opcional (cuerpo desplazable)
    Información del registro
    Cancelar | Guardar (pie fijo)

La revisión del patrón anterior encontró que una tabla encima de un formulario
permanente aleja la carga cuando crece el listado. Se sustituye por apertura
explícita del editor. La disposición visual coincide con el orden de teclado.

## Reglas de interacción

- Botones comunes: Nuevo, Editar, Eliminar, Cancelar y Guardar. El título identifica la entidad.
- Cada campo requerido lleva «Obligatorio» junto a la etiqueta, mediante `FormLabel`.
  El control conserva `required`. La indicación permanece visible al ocultar las ayudas.
  No usar una lista general de nombres de campos para informar obligatoriedad.
- Nuevo abre datos vacíos y enfoca Código; Editar carga únicamente campos editables.
- El listado conserva filtros y página al cerrar el editor.
- Paginación de 10, 25 o 50 filas. Altura acotada del listado, cabecera y acciones fijas.
- Filtrar antes de paginar; volver a página 1 al cambiar filtros o tamaño de página.
- Si desaparece la última página por eliminación, regresar a la última válida.
- Búsqueda general más filtros por campo combinados con AND. Mostrar cantidad y limpieza
  de filtros activos incluso con el panel de filtros cerrado.
- Editar usa botón con lápiz y texto. Eliminar usa botón con papelera, nombre accesible
  que identifica el registro y confirmación propia; las acciones quedan a la derecha.
- Un mismo estado controla ayuda general y de los campos. Se recuerda por pantalla
  en este navegador. El panel reproduce ese control para que se pueda usar al editar.
- Auditoría de solo lectura dentro del editor. Antes del primer guardado: «Sin guardar».
- Cancelar, Cerrar y Escape protegen cambios pendientes mediante diálogo propio.
- Foco contenido por el diálogo modal nativo estilizado y restaurado al activador al cerrar.
- Guardado fallido conserva los datos; el listado se actualiza solo después de persistir.

## Reutilización y límites

`web/src/ui/SmallFormDialog.vue` comparte apertura, foco, altura y pie fijo.
`web/src/ui/FormLabel.vue` comparte etiqueta vinculada al control e indicación de obligatoriedad.
`web/src/composables/usePagination.ts` comparte paginación de conjuntos locales completos.
`UnitsView.vue` conserva campos, validaciones y filtros propios del dominio.

La implementación sigue usando el catálogo local existente. Para grandes volúmenes
centrales, búsqueda, filtros, orden y paginación se ejecutarán en servidor y el cliente
recibirá solo una página y el total. Paginar localmente no reduce los datos descargados.
Esta entrega no implementa permisos centrales ni transforma metadatos locales históricos
en auditoría verificada del servidor.

Referencia de consulta: https://carbondesignsystem.com/components/data-table/usage/
(barra de herramientas, acciones por fila y paginación). Skill aplicado: frontend-design.
