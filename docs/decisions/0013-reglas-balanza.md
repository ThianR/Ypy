# ADR 0013 — Códigos de peso/precio

Las reglas de balanza se descargan como configuración versionada. Una coincidencia exacta tiene prioridad sobre prefijos; prefijos de igual prioridad son ambiguos y se rechazan. Peso o precio cero, formato incompleto y valores no numéricos no generan una línea POS.

