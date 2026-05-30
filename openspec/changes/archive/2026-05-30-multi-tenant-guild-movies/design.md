## Context

El bot usa SQLite (a través de `modernc.org/sqlite`) con una tabla `peliculas` que no incluye ningún identificador de servidor. Todos los handlers reciben `*discordgo.InteractionCreate`, cuyo campo `i.GuildID` provee el ID del guild (servidor) que originó la interacción. Actualmente ese campo se ignora.

## Goals / Non-Goals

**Goals:**
- Aislar el catálogo de películas por `guild_id` de forma que cada servidor opere sobre su propio conjunto de datos.
- Hacer la migración de la base de datos existente sin pérdida de registros.
- Mantener el comportamiento actual de todos los comandos sin cambios visibles para el usuario.

**Non-Goals:**
- Soporte para interacciones fuera de un guild (DMs).
- Sincronización o copia de catálogos entre servidores.
- Sistema de permisos por guild.

## Decisions

### 1. Añadir `guild_id` como columna en `peliculas`

**Decisión**: Agregar `guild_id TEXT NOT NULL DEFAULT ''` y luego hacer que todos los queries lo filtren.

**Alternativas consideradas**:
- *Base de datos separada por guild*: más aislamiento, pero complejidad operacional alta (múltiples archivos, conexiones dinámicas).
- *Prefijo en el nombre de la película*: hack frágil, no escalable.

**Rationale**: La columna adicional es el enfoque estándar de multi-tenancy en bases de datos relacionales. Es simple, consultas eficientes con índice compuesto, y no requiere cambios de dependencias.

### 2. Migración con `ALTER TABLE` + `DEFAULT ''`

**Decisión**: Usar `ALTER TABLE peliculas ADD COLUMN guild_id TEXT NOT NULL DEFAULT ''` para bases de datos existentes, detectando si la columna ya existe antes de ejecutarlo.

**Alternativas consideradas**:
- *DROP + recrear tabla*: destruye datos existentes.
- *Migración versionada*: mayor infraestructura de la que el proyecto necesita ahora.

**Rationale**: SQLite soporta `ADD COLUMN` con `DEFAULT` sin reconstruir la tabla. Detectar la existencia de la columna con `PRAGMA table_info` hace la operación idempotente.

### 3. Propagación de `guild_id` dentro de cada handler

**Decisión**: Extraer `i.GuildID` al inicio de cada handler y pasarlo directamente a las queries SQL.

**Alternativas consideradas**:
- *Middleware / wrapper*: añade indirección sin beneficio real para un bot con pocos handlers.

**Rationale**: El código ya es un archivo Go único. La propagación explícita es más legible y directa que introducir abstracciones.

### 4. Índice compuesto `(guild_id, nombre)`

**Decisión**: Añadir un índice `CREATE INDEX IF NOT EXISTS idx_peliculas_guild ON peliculas(guild_id)` para optimizar los filtros más frecuentes.

**Rationale**: Con múltiples guilds la tabla puede crecer significativamente; el índice evita full-scans por guild.

## Risks / Trade-offs

- **Registros legados con `guild_id = ''`** → Quedarán huérfanos e invisibles para todos los guilds. Riesgo bajo: ambiente de desarrollo; en producción la BD empieza vacía.
- **`i.GuildID` vacío en DMs** → Si el bot recibe un comando fuera de un guild el `guild_id` será `""`, mezclando datos DM de distintos usuarios. Mitigación: rechazar interacciones con `guild_id` vacío con un mensaje de error amigable.
- **SQLite en `/tmp`** → Los datos se pierden al reiniciar el sistema. Riesgo preexistente, fuera del alcance de este cambio.

## Migration Plan

1. Al arrancar `initDB()`, ejecutar el schema normal (`CREATE TABLE IF NOT EXISTS`).
2. Intentar `ALTER TABLE peliculas ADD COLUMN guild_id TEXT NOT NULL DEFAULT ''`; ignorar el error si la columna ya existe (SQLite devuelve `duplicate column name`).
3. Crear el índice con `CREATE INDEX IF NOT EXISTS`.
4. Desplegar el binario; no se requiere downtime ya que SQLite no bloquea en `ALTER TABLE ADD COLUMN`.

**Rollback**: Volver al binario anterior; los registros con `guild_id` poblado seguirán siendo accesibles (la columna extra es ignorada por el código viejo).
