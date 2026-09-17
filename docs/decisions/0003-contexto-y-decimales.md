# ADR 0003 — Contexto de empresa e importes

El contexto activo valida empresa y sucursal en cada operación. Los identificadores y valores monetarios cruzan contratos como cadenas; el núcleo Go usa `math/big` para evitar pérdida de precisión. La conversión histórica se incorporará al contrato de cotizaciones sin reescribir documentos confirmados.

