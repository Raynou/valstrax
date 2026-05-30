## 1. Migración de base de datos

- [x] 1.1 Modificar `initDB` para añadir la columna `guild_id TEXT NOT NULL DEFAULT ''` con `ALTER TABLE ... ADD COLUMN` (idempotente: ignorar error si ya existe)
- [x] 1.2 Añadir `CREATE INDEX IF NOT EXISTS idx_peliculas_guild ON peliculas(guild_id)` en `initDB`
- [x] 1.3 Actualizar el schema del `CREATE TABLE IF NOT EXISTS` para incluir `guild_id TEXT NOT NULL` (para bases de datos nuevas)

## 2. Guardia de interacciones sin guild

- [x] 2.1 Crear función helper `guildID(i *discordgo.InteractionCreate) (string, bool)` que retorne el `GuildID` y `false` si está vacío
- [x] 2.2 Añadir comprobación al inicio de cada handler: si `guild_id` está vacío, responder con mensaje de error y retornar

## 3. Actualizar handlers de escritura

- [x] 3.1 `handleAdd`: incluir `guild_id` en el `INSERT INTO peliculas`
- [x] 3.2 `handleRemove`: añadir `AND guild_id = ?` al `DELETE`
- [x] 3.3 `handleMarkMovieAsSeen`: añadir `AND guild_id = ?` al `SELECT id` y al `UPDATE`
- [x] 3.4 `handleUnmarkMovie`: añadir `AND guild_id = ?` al `UPDATE` (tanto por nombre como por id)

## 4. Actualizar handlers de lectura

- [x] 4.1 `handleSuggest`: añadir `AND guild_id = ?` al `SELECT`
- [x] 4.2 `handleGetMovieList`: añadir `AND guild_id = ?` al `SELECT`

## 5. Actualizar autocompletado

- [x] 5.1 `respondAutocomplete`: añadir `guild_id` como parámetro y filtrar el `SELECT` por `guild_id`
- [x] 5.2 `respondAutocompleteSeenMovies`: añadir `guild_id` como parámetro y filtrar el `SELECT` por `guild_id`
- [x] 5.3 Actualizar todas las llamadas a `respondAutocomplete` y `respondAutocompleteSeenMovies` para pasar `i.GuildID`

## 6. Verificación

- [x] 6.1 Compilar el proyecto (`go build ./...`) y corregir errores
- [x] 6.2 Verificar manualmente que dos sesiones con distintos `GUILD_ID` no comparten películas
