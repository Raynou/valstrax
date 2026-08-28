## Why

El bot corre como un proceso permanente conectado al gateway de Discord por WebSocket y guarda su catálogo en `/tmp/peliculas.db`, una ruta efímera. Eso obliga a mantener una máquina encendida 24/7 y expone los datos a perderse en cualquier reinicio. Como el bot solo expone slash commands y nunca consume eventos del gateway, la conexión permanente no aporta nada: el modelo request/response de Cloudflare Workers cubre el caso de uso completo, elimina el servidor a mantener y entra de sobra en la capa gratuita.

## What Changes

- **BREAKING**: Se sustituye la conexión gateway WebSocket (`dg.Open()`) por un Interactions Endpoint HTTP. El bot deja de recibir eventos del gateway (mensajes, reacciones, estado de voz) y solo atiende invocaciones explícitas de slash commands.
- **BREAKING**: El worker se reescribe en TypeScript sobre Cloudflare Workers. Se retira el binario Go y sus dependencias (`discordgo`, `modernc.org/sqlite`).
- Se añade verificación de firma Ed25519 en cada petición entrante y respuesta `PONG` al `PING` de validación de Discord.
- La persistencia pasa de SQLite local a Cloudflare D1. El SQL existente se conserva casi íntegro por seguir siendo SQLite.
- Se portan los 6 slash commands (`agregar`, `sugerir`, `vista`, `lista`, `quitar`, `desmarcar`) y los 2 flujos de autocompletado, preservando el filtrado por `guild_id` en todas las consultas.
- El registro de comandos deja de ocurrir al arrancar el proceso y pasa a ser un script one-off ejecutable a mano o desde CI.
- Se migra el esquema y los datos existentes de la tabla `peliculas` a D1.

## Capabilities

### New Capabilities

- `http-interactions`: Recepción de interacciones de Discord por HTTP — verificación de firma Ed25519, manejo del handshake `PING`/`PONG`, rechazo de peticiones no firmadas y respuesta dentro de la ventana de 3 segundos que exige Discord.
- `catalog-persistence`: Durabilidad del catálogo entre invocaciones y despliegues. Ninguna invocación puede depender de estado en memoria o en disco local.

### Modified Capabilities

*(ninguna: `guild-scoped-movies` conserva sus requisitos intactos — el aislamiento por servidor se mantiene idéntico, solo cambia el motor de persistencia que lo implementa)*

## Impact

- **`main.go`**: Se retira por completo. Su lógica de handlers y consultas se porta al worker en TypeScript.
- **`go.mod` / `go.sum`**: Se eliminan junto con el resto del proyecto Go.
- **Nuevas dependencias**: `wrangler` (build y despliegue), `discord-interactions` o WebCrypto nativo para la verificación de firma.
- **Nueva configuración**: `wrangler.toml` con el binding de D1; los secretos `DISCORD_TOKEN`, `DISCORD_PUBLIC_KEY` y `DISCORD_APPLICATION_ID` pasan a secretos de Workers.
- **Discord Developer Portal**: Hay que fijar el Interactions Endpoint URL, que Discord valida con una petición firmada antes de aceptarlo.
- **Datos**: La base en `/tmp/peliculas.db` solo eran datos de prueba, por lo que se puede descartar la migracion de datos de SQLite a D1.
- **Límites operativos**: Workers Free (100k req/día, 10 ms CPU por petición, bundle de 3 MB comprimido) y D1 Free (5M filas leídas/día, 100k escritas/día, 50 consultas por invocación) quedan muy por encima del volumen esperado.
