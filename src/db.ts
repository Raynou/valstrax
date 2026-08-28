// Acceso a datos sobre D1. Toda consulta filtra por guild_id: el aislamiento
// por servidor es un requisito de la capability guild-scoped-movies y no debe
// relajarse aquí.

export interface MovieListRow {
	nombre: string;
	vista: number;
}

/** Selector opcional por nombre o por id, tal como lo aceptan /quitar y /desmarcar. */
export interface MovieSelector {
	nombre: string;
	id: number;
}

export async function addMovie(
	db: D1Database,
	guildID: string,
	nombre: string,
	agregadaPor: string,
): Promise<number> {
	const result = await db
		.prepare(
			`INSERT INTO peliculas (nombre, agregada_por, guild_id) VALUES (?, ?, ?)`,
		)
		.bind(nombre, agregadaPor, guildID)
		.run();
	return result.meta.last_row_id;
}

export async function randomUnseenMovie(
	db: D1Database,
	guildID: string,
): Promise<string | null> {
	const row = await db
		.prepare(
			`SELECT nombre FROM peliculas
			 WHERE vista_en IS NULL AND guild_id = ?
			 ORDER BY RANDOM()
			 LIMIT 1`,
		)
		.bind(guildID)
		.first<{ nombre: string }>();
	return row?.nombre ?? null;
}

export async function findMovieIDByName(
	db: D1Database,
	guildID: string,
	nombre: string,
): Promise<number | null> {
	const row = await db
		.prepare(
			`SELECT id FROM peliculas WHERE nombre = ? COLLATE NOCASE AND guild_id = ?`,
		)
		.bind(nombre, guildID)
		.first<{ id: number }>();
	return row?.id ?? null;
}

export async function markMovieSeen(
	db: D1Database,
	guildID: string,
	id: number,
): Promise<void> {
	await db
		.prepare(
			`UPDATE peliculas SET vista_en = CURRENT_TIMESTAMP WHERE id = ? AND guild_id = ?`,
		)
		.bind(id, guildID)
		.run();
}

export async function listMovies(
	db: D1Database,
	guildID: string,
): Promise<MovieListRow[]> {
	// TODO heredado: sin paginación, se corta en 25. Fuera del alcance de este cambio.
	const result = await db
		.prepare(
			`SELECT nombre,
			        CASE WHEN vista_en IS NOT NULL THEN 1 ELSE 0 END AS vista
			 FROM peliculas
			 WHERE guild_id = ?
			 LIMIT 25`,
		)
		.bind(guildID)
		.all<MovieListRow>();
	return result.results;
}

/**
 * Construye la condición de un selector por nombre y/o id. Si se dan ambos,
 * coincide con cualquiera de los dos, igual que la implementación anterior.
 */
function selectorClause(
	selector: MovieSelector,
): { sql: string; binds: unknown[] } | null {
	const clauses: string[] = [];
	const binds: unknown[] = [];

	if (selector.nombre !== "") {
		clauses.push("nombre = ? COLLATE NOCASE");
		binds.push(selector.nombre);
	}
	if (selector.id !== 0) {
		clauses.push("id = ?");
		binds.push(selector.id);
	}
	if (clauses.length === 0) return null;

	return { sql: `(${clauses.join(" OR ")})`, binds };
}

/** Devuelve el número de películas eliminadas. */
export async function removeMovies(
	db: D1Database,
	guildID: string,
	selector: MovieSelector,
): Promise<number> {
	const clause = selectorClause(selector);
	if (clause === null) return 0;

	const result = await db
		.prepare(`DELETE FROM peliculas WHERE guild_id = ? AND ${clause.sql}`)
		.bind(guildID, ...clause.binds)
		.run();
	return result.meta.changes;
}

/** Devuelve el número de películas marcadas como no vistas. */
export async function unmarkMovies(
	db: D1Database,
	guildID: string,
	selector: MovieSelector,
): Promise<number> {
	const clause = selectorClause(selector);
	if (clause === null) return 0;

	const result = await db
		.prepare(
			`UPDATE peliculas SET vista_en = NULL WHERE guild_id = ? AND ${clause.sql}`,
		)
		.bind(guildID, ...clause.binds)
		.run();
	return result.meta.changes;
}

/**
 * Nombres que coinciden con lo escrito, filtrados por estado de visto.
 * Discord admite como mucho 25 sugerencias.
 */
export async function searchMovieNames(
	db: D1Database,
	guildID: string,
	query: string,
	seen: boolean,
): Promise<string[]> {
	const estado = seen ? "vista_en IS NOT NULL" : "vista_en IS NULL";
	const result = await db
		.prepare(
			`SELECT nombre FROM peliculas
			 WHERE nombre LIKE ? COLLATE NOCASE
			   AND ${estado}
			   AND guild_id = ?
			 LIMIT 25`,
		)
		.bind(`%${query}%`, guildID)
		.all<{ nombre: string }>();
	return result.results.map((row) => row.nombre);
}
