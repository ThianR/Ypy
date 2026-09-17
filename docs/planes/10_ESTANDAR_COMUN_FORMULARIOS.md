# Estándar común de formularios de Ypy

## Alcance y estado

Una base común de interacción con variantes de distribución según la tarea.
Implementado como piloto en Unidades de medida: etiquetas de obligatoriedad por campo,
ayuda conmutable, panel de edición, pie de acciones, protección de cambios y paginación local.
Las reglas para relaciones y maestro-detalle de este documento son criterios de diseño
para próximas entregas; no acreditan componentes ni pantallas ya implementados.

## Reglas comunes

- Etiqueta visible vinculada al control mediante un identificador único.
- Cada campo requerido muestra «Obligatorio» al lado de su nombre. El texto evita depender
  del color o de conocer el significado de un asterisco. Permanece visible con ayudas ocultas.
- `FormLabel.vue` presenta la etiqueta y su marca a partir de `required`. El control nativo
  debe recibir el mismo estado `required`; los controles compuestos exponen `aria-required`
  y validación equivalente. La obligatoriedad de negocio también se valida en servidor.
- Un campo condicional cambia su marca y su validación con la misma condición de negocio.
  No exigir campos ocultos que el usuario no pueda completar; resolverlos o mostrar la sección.
- Ayuda opcional explica significado y ejemplos. Etiquetas, obligatoriedad y errores no se ocultan.
- Errores junto al campo, sin borrar datos. En formularios extensos, resumen con enlaces
  al campo o fila afectada y apertura de la sección correspondiente antes de enfocar.
- Acciones consistentes: Nuevo, Editar, Eliminar, Cancelar y Guardar. Confirmar operaciones
  comerciales es una acción específica distinta de guardar un borrador cuando el dominio lo requiera.
- Auditoría de solo lectura, cambios pendientes protegidos y errores de guardado recuperables.

## Variantes de distribución

| Variante | Uso | Distribución prevista |
| --- | --- | --- |
| Pequeño | Unidad de medida, catálogo con pocos campos | Listado paginado y panel lateral de carga/edición |
| Con secciones y relaciones | Artículo, cliente o proveedor con múltiples grupos de datos | Página de edición amplia, secciones y selectores con búsqueda; acciones persistentes |
| Maestro-detalle | Pedido, compra u otro documento con cabecera y renglones | Página de trabajo: cabecera, detalle editable, totales y acciones persistentes |

El tamaño y la complejidad determinan la distribución. Tener una sola relación no obliga
a abandonar el panel pequeño. No colocar un documento con muchas líneas dentro de un panel estrecho.

## Relaciones

El usuario selecciona por código y descripción reconocibles, no por un identificador interno.
La etiqueta del selector muestra su obligatoriedad. La búsqueda debe indicar carga, ausencia
de resultados y errores. Al cambiar un dato padre, revisar los valores dependientes y explicar
cualquier selección que deje de ser válida. Crear un relacionado, si los permisos lo permiten,
debe conservar lo ya escrito en el formulario de origen.

## Maestro-detalle

La cabecera mantiene las mismas etiquetas. Cada columna requerida del detalle editable
indica «Obligatorio»; el editor de una fila repite esa identificación en sus campos.
En móvil las filas pueden editarse como fichas sin perder etiqueta, obligatoriedad ni errores.
Identificar errores por fila y campo, por ejemplo «Fila 3, Artículo: selecciona un artículo».

La regla «debe existir al menos una línea» pertenece al detalle completo y se comunica allí,
además de la obligatoriedad de las celdas. No marcar un checkbox como requerido cuando
`false` sea una respuesta válida; vacío, cero y falso son valores distintos.
Validar todas las líneas, incluidas las no visibles por paginación. El guardado del documento
debe respetar la transacción y las reglas del negocio; una fila visible no equivale al documento completo.

## Crecimiento de los componentes

Reutilizar etiquetas, campos, mensajes, búsqueda, auditoría y acciones con la misma semántica.
Mantener las reglas de cada entidad en su módulo. Incorporar componentes de relación y detalle
al implementar casos reales, sin convertir el formulario pequeño en un componente con todas las variantes.
