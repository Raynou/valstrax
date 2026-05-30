## ADDED Requirements

### Requirement: Aislamiento de catálogo por servidor
Cada servidor de Discord (guild) SHALL tener su propio catálogo de películas independiente. Las operaciones de lectura y escritura DEBEN estar siempre filtradas por el `guild_id` del servidor donde se originó la interacción.

#### Scenario: Agregar película en un servidor no afecta a otro
- **WHEN** un usuario del servidor A ejecuta `/agregar` con el nombre de una película
- **THEN** la película se almacena asociada al `guild_id` del servidor A y no aparece en el catálogo del servidor B

#### Scenario: Listar películas muestra solo las del servidor actual
- **WHEN** un usuario ejecuta `/lista` en el servidor B
- **THEN** el bot responde únicamente con las películas agregadas en el servidor B

#### Scenario: Sugerir película respeta el catálogo del servidor
- **WHEN** un usuario ejecuta `/sugerir` en un servidor
- **THEN** el bot sugiere una película aleatoria no vista del catálogo de ese servidor, sin incluir películas de otros servidores

### Requirement: Rechazo de interacciones sin contexto de servidor
El bot SHALL rechazar cualquier interacción que no provenga de un servidor de Discord (e.g., mensajes directos), respondiendo con un mensaje de error claro.

#### Scenario: Comando ejecutado en DM
- **WHEN** un usuario intenta ejecutar cualquier comando del bot en un mensaje directo (sin `guild_id`)
- **THEN** el bot responde con un mensaje indicando que los comandos solo están disponibles dentro de un servidor

### Requirement: Persistencia correcta del guild_id en todas las operaciones
Todas las operaciones de escritura (agregar, marcar como vista, desmarcar, eliminar) SHALL almacenar el `guild_id` de la interacción junto con el registro, y todas las operaciones de lectura SHALL filtrar por ese mismo `guild_id`.

#### Scenario: Marcar película como vista aplica solo dentro del servidor
- **WHEN** un usuario del servidor C marca una película como vista con `/vista`
- **THEN** el estado `vista_en` se actualiza únicamente para el registro de esa película en el servidor C

#### Scenario: Eliminar película con `/quitar` opera sobre el catálogo del servidor
- **WHEN** un usuario elimina una película por nombre o id en el servidor D
- **THEN** solo se eliminan registros con el `guild_id` del servidor D, sin tocar películas de otros servidores

#### Scenario: Autocompletado filtra por servidor
- **WHEN** un usuario escribe en el campo de autocompletado de `/vista`, `/quitar` o `/desmarcar`
- **THEN** las sugerencias mostradas corresponden únicamente a películas del servidor actual
