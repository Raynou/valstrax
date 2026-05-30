## Why

La tabla `peliculas` no tiene ninguna columna que identifique el servidor (guild) de Discord, por lo que todas las instancias del bot comparten una única lista global de películas. Esto hace imposible desplegar el bot en múltiples servidores sin que sus catálogos se mezclen.

## What Changes

- Se añade la columna `guild_id TEXT NOT NULL` a la tabla `peliculas`.
- Todas las consultas SQL (INSERT, SELECT, UPDATE, DELETE) se filtran por `guild_id`.
- El `guild_id` se extrae del campo `i.GuildID` de cada interacción de Discord.
- Se hace una migración de la base de datos existente para añadir la columna con un valor por defecto vacío para registros legados (si los hay).

## Capabilities

### New Capabilities

- `guild-scoped-movies`: Aislamiento de datos por servidor; cada guild tiene su propio catálogo de películas independiente.

### Modified Capabilities

*(ninguna: los requisitos a nivel de spec no cambian, solo la implementación subyacente de persistencia)*

## Impact

- **`main.go`**: Todos los handlers (`handleAdd`, `handleRemove`, `handleSuggest`, `handleMarkMovieAsSeen`, `handleGetMovieList`, `handleUnmarkMovie`) y las funciones de autocompletado (`respondAutocomplete`, `respondAutocompleteSeenMovies`) necesitan recibir y propagar el `guild_id`.
- **`initDB`**: El schema de la tabla `peliculas` debe incluir `guild_id`; se necesita migración para bases de datos existentes.
- **Sin nuevas dependencias**: No se requieren librerías adicionales.
