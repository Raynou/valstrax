-- Esquema del catálogo de películas.
-- Reproduce el esquema del SQLite anterior: mismas columnas y mismo índice por guild.

CREATE TABLE IF NOT EXISTS peliculas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nombre TEXT NOT NULL,
    agregada_por TEXT NOT NULL,
    guild_id TEXT NOT NULL DEFAULT '',
    creada_en DATETIME DEFAULT CURRENT_TIMESTAMP,
    vista_en DATETIME
);

CREATE INDEX IF NOT EXISTS idx_peliculas_guild ON peliculas(guild_id);
