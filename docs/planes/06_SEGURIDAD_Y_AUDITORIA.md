# Seguridad y auditoría funcional

## Autorización

Identidad resuelve usuario, empresa, sucursal, terminal y sesión. Los roles agrupan permisos; los permisos se evalúan sobre la operación y su ámbito. Preparar ámbitos de depósito/caja además de empresa/sucursal, sin conceder automáticamente acceso a todos los depósitos por pertenecer a la empresa. Revisar el modelo V4 y agregar solo las relaciones faltantes.

La terminal no constituye por sí sola la identidad del operador. En cada operación contrastar usuario, permiso, sesión de caja y terminal. No trasladar terminal de sucursal mientras una sesión siga activa. Una habilitación técnica de funcionalidad tampoco concede permisos ni autorización fiscal.

Cuenta de aplicación con privilegios mínimos, distinta a migraciones. Acceso central por HTTPS. Sesión del navegador con protección frente a robo/uso cruzado y CSRF cuando use cookies; agente local con enrolamiento, orígenes permitidos y autenticación. Guardar claves del dispositivo con mecanismos del sistema cuando estén disponibles; nunca incluir claves privadas de firma central en las cajas. Los permisos offline contienen evidencia verificable, no autoridad para emitir nuevos permisos libremente.

## Auditoría de negocio

Reutilizar `gs_auditoria`, que ya conserva empresa, usuario, fecha, tabla/registro, acción, antes y después. F06 verifica cobertura actual y agrega por migración los campos necesarios: UUID de operación, terminal/instalación, motivo, request_id y correlation_id. No crear otra tabla genérica de auditoría sin una necesidad concreta.

Auditar cambios de precio/cotización/configuración, permisos, numeración, ajustes de cantidad/costo, devoluciones, anulaciones, ingresos/retiros de caja, cierre y resolución de conflictos. Operaciones sensibles requieren motivo y permiso específico. Reapertura de caja no queda habilitada por mencionar auditoría: fuera del circuito inicial, corregir mediante procedimientos trazables.

Para cambios confirmados, auditoría y operación se guardan en la misma transacción; si falla la escritura requerida, no confirmar a medias. Para intentos rechazados o fallas de autenticación, registrar evento de seguridad por separado para que el rollback no borre la evidencia del intento. Offline, guardar evidencia local con la operación y verificarla al integrar.

La aplicación no puede modificar/eliminar evidencia ya confirmada. Esto no implica resistencia absoluta contra un administrador de base: definir retención, exportación y acceso de soporte explícitamente. Antes/después contienen solo los campos necesarios; excluir passwords, claves, tokens y datos de tarjeta. Distinguir auditoría funcional, kardex y logs de diagnóstico: tienen propósitos y retenciones diferentes.

## Configuración gradual

Reutilizar `gs_empresa_modulo` y sus opciones para habilitaciones que encajen allí. Documentar opciones tipadas, ámbito, valor por defecto y versión del snapshot offline. Revisar necesidad antes de agregar feature flags genéricos.

Desactivar una función impide nuevas operaciones según política, pero no debe descartar automáticamente operaciones históricas válidas creadas con permiso/configuración anterior. Los casos de revocación requieren resolución explícita. El indicador de venta negativa solo funciona junto con permiso y origen POS autenticado.

## Aceptación

F01/F06 y T26 verifican cambios de ámbito, autorización de supervisor, auditoría atómica, redacción de secretos y configuración que no eluda permisos. Los logs incluyen request_id por solicitud y correlation_id/operation_id estable entre reintentos, integración y fiscalización.
