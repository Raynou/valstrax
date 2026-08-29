// Registro de los slash commands en Discord.
//
// Script one-off: ejecútalo solo cuando cambie la definición de los comandos,
// no en cada despliegue. El worker no registra nada al arrancar porque no hay
// arranque que aprovechar.
//
//   npm run register
//
// Lee DISCORD_TOKEN, DISCORD_APPLICATION_ID y, opcionalmente, GUILD_ID del
// entorno. En local se cargan desde .env (`--env-file-if-exists`); en la
// pipeline de CI se inyectan desde GitHub Actions Secrets. Con GUILD_ID los
// comandos se registran solo en ese servidor y aparecen al instante — útil para
// probar. Sin él, se registran globalmente y Discord puede tardar en propagarlos.

export {}; // marca el fichero como módulo, para poder usar await de nivel superior

const OptionType = {
	STRING: 3,
	INTEGER: 4,
} as const;

const commands = [
	{
		name: "agregar",
		description: "Agrega una película al catálogo",
		options: [
			{
				type: OptionType.STRING,
				name: "nombre",
				description: "Nombre de la película",
				required: true,
			},
		],
	},
	{
		name: "sugerir",
		description: "Sugiere una película que aún no hayas visto",
	},
	{
		name: "vista",
		description: "Marca una película como vista por ti",
		options: [
			{
				type: OptionType.STRING,
				name: "nombre",
				description: "Película a marcar como vista",
				required: true,
				autocomplete: true,
			},
		],
	},
	{
		name: "lista",
		description: "Muestra el catálogo de películas",
	},
	{
		name: "quitar",
		description: "Elimina una pelicula de la lista de pendientes",
		options: [
			{
				type: OptionType.STRING,
				name: "nombre",
				description: "Nombre de la película a eliminar",
				required: false,
				autocomplete: true,
			},
			{
				type: OptionType.INTEGER,
				name: "id",
				description: "ID de la película a eliminar",
				required: false,
			},
		],
	},
	{
		name: "desmarcar",
		description: "Marca una pelicula como no vista",
		options: [
			{
				type: OptionType.STRING,
				name: "nombre",
				description: "Nombre de la pelicula",
				required: false,
				autocomplete: true,
			},
			{
				type: OptionType.INTEGER,
				name: "id",
				description: "Id de la pelicula",
				required: false,
			},
		],
	},
];

const token = process.env.DISCORD_TOKEN;
const applicationID = process.env.DISCORD_APPLICATION_ID;
const guildID = process.env.GUILD_ID;

if (!token || !applicationID) {
	console.error(
		"Faltan DISCORD_TOKEN o DISCORD_APPLICATION_ID en el entorno (.env).",
	);
	process.exit(1);
}

const url = guildID
	? `https://discord.com/api/v10/applications/${applicationID}/guilds/${guildID}/commands`
	: `https://discord.com/api/v10/applications/${applicationID}/commands`;

const response = await fetch(url, {
	method: "PUT",
	headers: {
		authorization: `Bot ${token}`,
		"content-type": "application/json",
	},
	body: JSON.stringify(commands),
});

if (!response.ok) {
	console.error(`Error ${response.status}: ${await response.text()}`);
	process.exit(1);
}

const registered = (await response.json()) as unknown[];
const alcance = guildID ? `en el servidor ${guildID}` : "globalmente";
console.log(`${registered.length} comandos registrados ${alcance}.`);
