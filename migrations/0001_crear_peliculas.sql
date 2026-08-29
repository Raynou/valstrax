-- Migración inicial del catálogo de películas.
-- Copia exacta del antiguo schema.sql. Los IF NOT EXISTS la hacen un no-op sobre la
-- base D1 remota, que ya tenía la tabla y el índice del despliegue manual previo (ver
-- design.md D2: adopción del sistema de migraciones sobre la base ya provisionada).

CREATE TABLE IF NOT EXISTS peliculas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nombre TEXT NOT NULL,
    agregada_por TEXT NOT NULL,
    guild_id TEXT NOT NULL DEFAULT '',
    creada_en DATETIME DEFAULT CURRENT_TIMESTAMP,
    vista_en DATETIME
);

CREATE INDEX IF NOT EXISTS idx_peliculas_guild ON peliculas(guild_id);
