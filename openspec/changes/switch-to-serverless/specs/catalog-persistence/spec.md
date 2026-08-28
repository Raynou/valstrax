## ADDED Requirements

### Requirement: Durabilidad del catálogo entre invocaciones

El catálogo de películas SHALL persistir en un almacenamiento externo al ciclo de vida de la invocación. Ninguna operación puede depender de estado guardado en memoria o en el sistema de ficheros local del worker, dado que ambos se descartan entre invocaciones.

#### Scenario: Una película agregada sobrevive a invocaciones posteriores

- **WHEN** un usuario ejecuta `/agregar` y más tarde otro usuario ejecuta `/lista` en el mismo servidor
- **THEN** la película agregada aparece en el listado, aunque las dos interacciones hayan sido atendidas por invocaciones distintas del worker

#### Scenario: El catálogo sobrevive a un despliegue

- **WHEN** se despliega una nueva versión del worker
- **THEN** el catálogo previo permanece íntegro y accesible

#### Scenario: Sin escrituras a disco local

- **WHEN** cualquier handler necesita leer o escribir datos
- **THEN** la operación se dirige a la base de datos externa, y el worker no abre ningún fichero local

#### Scenario: Arranque en vacío

- **WHEN** se aplica el esquema por primera vez sobre la base D1
- **THEN** el catálogo arranca vacío y el bot opera con normalidad desde el primer `/agregar`

### Requirement: Preservación del aislamiento por servidor

Todas las consultas contra el almacenamiento externo SHALL mantener el filtrado por `guild_id` definido en la capability `guild-scoped-movies`. El cambio de motor de persistencia NO DEBE relajar ese aislamiento.

#### Scenario: El filtrado por servidor se mantiene tras el cambio de motor

- **WHEN** un usuario ejecuta cualquier comando de lectura o escritura sobre D1
- **THEN** la consulta resultante filtra por el `guild_id` de la interacción, igual que antes del cambio
