import { handleAutocomplete, handlers } from "./handlers.ts";
import {
	autocomplete,
	InteractionType,
	message,
	pong,
	type Interaction,
} from "./discord.ts";
import { verifyRequest } from "./verify.ts";

interface Env {
	DB: D1Database;
	DISCORD_PUBLIC_KEY: string;
}

const SIN_SERVIDOR = "Este comando solo está disponible dentro de un servidor.";

export default {
	async fetch(request: Request, env: Env): Promise<Response> {
		if (request.method !== "POST") {
			return new Response("Method not allowed", { status: 405 });
		}

		// La firma se verifica sobre el cuerpo crudo, antes de parsear nada.
		const body = await verifyRequest(request, env.DISCORD_PUBLIC_KEY);
		if (body === null) {
			return new Response("Bad request signature", { status: 401 });
		}

		let interaction: Interaction;
		try {
			interaction = JSON.parse(body) as Interaction;
		} catch {
			return new Response("Invalid JSON", { status: 400 });
		}

		// Handshake con el que Discord valida el Interactions Endpoint URL.
		if (interaction.type === InteractionType.PING) {
			return pong();
		}

		const guildID = interaction.guild_id;

		if (interaction.type === InteractionType.APPLICATION_COMMAND_AUTOCOMPLETE) {
			// Un autocompletado no admite respuesta de texto: fuera de un
			// servidor simplemente no hay sugerencias que ofrecer.
			if (!guildID) return autocomplete([]);
			return handleAutocomplete(interaction, guildID, env.DB);
		}

		if (interaction.type !== InteractionType.APPLICATION_COMMAND) {
			return message("Tipo de interacción no soportado.");
		}

		if (!guildID) return message(SIN_SERVIDOR);

		const handler = handlers[interaction.data?.name ?? ""];
		if (handler === undefined) {
			return message("Comando no reconocido.");
		}

		return handler(interaction, guildID, env.DB);
	},
} satisfies ExportedHandler<Env>;
