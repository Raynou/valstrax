// Tipos y helpers mínimos del protocolo de interacciones de Discord.
// Solo se modela lo que este bot usa; no es un cliente general.

export const InteractionType = {
	PING: 1,
	APPLICATION_COMMAND: 2,
	APPLICATION_COMMAND_AUTOCOMPLETE: 4,
} as const;

export const InteractionResponseType = {
	PONG: 1,
	CHANNEL_MESSAGE_WITH_SOURCE: 4,
	APPLICATION_COMMAND_AUTOCOMPLETE_RESULT: 8,
} as const;

export interface InteractionOption {
	name: string;
	type: number;
	value?: string | number | boolean;
	focused?: boolean;
}

export interface Interaction {
	type: number;
	guild_id?: string;
	member?: { user: { id: string } };
	user?: { id: string };
	data?: {
		name: string;
		options?: InteractionOption[];
	};
}

export interface Choice {
	name: string;
	value: string;
}

function json(body: unknown): Response {
	return new Response(JSON.stringify(body), {
		headers: { "content-type": "application/json" },
	});
}

export function pong(): Response {
	return json({ type: InteractionResponseType.PONG });
}

export function message(content: string): Response {
	return json({
		type: InteractionResponseType.CHANNEL_MESSAGE_WITH_SOURCE,
		data: { content },
	});
}

export function autocomplete(choices: Choice[]): Response {
	return json({
		type: InteractionResponseType.APPLICATION_COMMAND_AUTOCOMPLETE_RESULT,
		data: { choices },
	});
}

export function optionValue(
	interaction: Interaction,
	name: string,
): string | number | undefined {
	const option = interaction.data?.options?.find((o) => o.name === name);
	if (option === undefined || typeof option.value === "boolean") return undefined;
	return option.value;
}

export function stringOption(interaction: Interaction, name: string): string {
	const value = optionValue(interaction, name);
	return typeof value === "string" ? value : "";
}

export function integerOption(interaction: Interaction, name: string): number {
	const value = optionValue(interaction, name);
	return typeof value === "number" ? value : 0;
}

/** Texto que el usuario lleva escrito en el campo enfocado de un autocompletado. */
export function focusedValue(interaction: Interaction): string {
	const option = interaction.data?.options?.find((o) => o.focused);
	return typeof option?.value === "string" ? option.value : "";
}

/** Id del usuario que invocó la interacción, dentro o fuera de un servidor. */
export function invokerID(interaction: Interaction): string {
	return interaction.member?.user.id ?? interaction.user?.id ?? "";
}
