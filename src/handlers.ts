// Handlers de los slash commands. Los mensajes al usuario se conservan
// literalmente respecto a la implementación anterior en Go.

import * as db from "./db.ts";
import {
	autocomplete,
	focusedValue,
	integerOption,
	invokerID,
	message,
	stringOption,
	type Choice,
	type Interaction,
} from "./discord.ts";

export type CommandHandler = (
	interaction: Interaction,
	guildID: string,
	database: D1Database,
) => Promise<Response>;

async function handleAdd(
	interaction: Interaction,
	guildID: string,
	database: D1Database,
): Promise<Response> {
	const nombre = stringOption(interaction, "nombre");
	const id = await db.addMovie(database, guildID, nombre, invokerID(interaction));
	return message(`✅ Agregada: **${nombre}** (id ${id})`);
}

async function handleSuggest(
	_interaction: Interaction,
	guildID: string,
	database: D1Database,
): Promise<Response> {
	const nombre = await db.randomUnseenMovie(database, guildID);
	if (nombre === null) {
		return message(
			"🎬 ¡Ya viste todo lo que hay en el catálogo! Agrega más con `/agregar`.",
		);
	}
	return message(
		`🎲 Sugerencia: **${nombre}**\nMárcala como vista con \`/vista\` cuando termines.`,
	);
}

async function handleMarkSeen(
	interaction: Interaction,
	guildID: string,
	database: D1Database,
): Promise<Response> {
	const nombre = stringOption(interaction, "nombre");
	const id = await db.findMovieIDByName(database, guildID, nombre);
	if (id === null) {
		return message(`No encontré **${nombre}** en el catálogo.`);
	}
	await db.markMovieSeen(database, guildID, id);
	return message(`👁️ Marcaste **${nombre}** como vista.`);
}

async function handleList(
	_interaction: Interaction,
	guildID: string,
	database: D1Database,
): Promise<Response> {
	const peliculas = await db.listMovies(database, guildID);
	if (peliculas.length === 0) {
		return message("El catálogo está vacío. Agrega algo con `/agregar`.");
	}
	const lineas = peliculas.map(
		(pelicula) => `${pelicula.vista === 1 ? "✅" : "⬜"} ${pelicula.nombre}`,
	);
	return message(`**Catálogo:**\n${lineas.join("\n")}\n`);
}

async function handleRemove(
	interaction: Interaction,
	guildID: string,
	database: D1Database,
): Promise<Response> {
	const selector = {
		nombre: stringOption(interaction, "nombre"),
		id: integerOption(interaction, "id"),
	};
	if (selector.nombre === "" && selector.id === 0) {
		return message("Debes proporcionar `nombre` o `id`.");
	}

	const eliminadas = await db.removeMovies(database, guildID, selector);
	if (eliminadas === 0) {
		return message("No encontre ninguna pelicula con esos datos");
	}
	return message(`📤 Eliminada(s) ${eliminadas} película(s).`);
}

async function handleUnmark(
	interaction: Interaction,
	guildID: string,
	database: D1Database,
): Promise<Response> {
	const selector = {
		nombre: stringOption(interaction, "nombre"),
		id: integerOption(interaction, "id"),
	};
	if (selector.nombre === "" && selector.id === 0) {
		return message("Debes proporcionar `nombre` o `id`.");
	}

	const desmarcadas = await db.unmarkMovies(database, guildID, selector);

	if (selector.nombre !== "") {
		if (desmarcadas === 0) {
			return message(
				`No encontre ninguna pelicula con el nombre: ${selector.nombre}`,
			);
		}
		return message(`Pelicula ${selector.nombre} marcada como no vista`);
	}

	if (desmarcadas === 0) {
		return message(`No encontré ninguna pelicula con la id: \`${selector.id}\``);
	}
	return message(`Pelicula con id \`${selector.id}\` marcada como no vista`);
}

export const handlers: Record<string, CommandHandler> = {
	agregar: handleAdd,
	sugerir: handleSuggest,
	vista: handleMarkSeen,
	lista: handleList,
	quitar: handleRemove,
	desmarcar: handleUnmark,
};

/**
 * Qué conjunto de películas sugiere el autocompletado de cada comando:
 * las pendientes para marcar o quitar, las ya vistas para desmarcar.
 */
const autocompleteSeen: Record<string, boolean> = {
	vista: false,
	quitar: false,
	desmarcar: true,
};

export async function handleAutocomplete(
	interaction: Interaction,
	guildID: string,
	database: D1Database,
): Promise<Response> {
	const comando = interaction.data?.name ?? "";
	const seen = autocompleteSeen[comando];
	if (seen === undefined) return autocomplete([]);

	const nombres = await db.searchMovieNames(
		database,
		guildID,
		focusedValue(interaction),
		seen,
	);
	const choices: Choice[] = nombres.map((nombre) => ({
		name: nombre,
		value: nombre,
	}));
	return autocomplete(choices);
}
