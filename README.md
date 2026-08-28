# Valstrax

Simple discord bot for managing movies to watch in your Discord server!

### Commands

#### Add movies

```bash
/agregar <name>
```

#### List added movies

```bash
/listar
```

#### Remove a movie from list

```bash
/quitar <name|id>
```

#### Suggest a movie

```bash
/sugerir
```

This retrieves a random movie that you've added before


#### Mark movie as seen

```bash
/vista <name|id>
```

#### Unmark movie as seen

```bash
/desmarcar <name|id>
```

### Development

The bot runs as a Cloudflare Worker backed by a D1 database.

Local secrets live in two separate files, because two different runtimes are
involved:

- `.dev.vars` — read by the Worker sandbox during `wrangler dev`. Needs
  `DISCORD_PUBLIC_KEY` (Discord Developer Portal → General Information).
- `.env` — read by Node when registering slash commands. Needs
  `DISCORD_TOKEN` and `DISCORD_APPLICATION_ID`, plus an optional `GUILD_ID`
  to register commands in a single server instead of globally.

```bash
npm install
npm run schema:local   # apply schema.sql to the local D1 database
npm run dev
```

### Deployment

```bash
npm run deploy
npx wrangler secret put DISCORD_PUBLIC_KEY
npm run schema:remote
npm run register
```

Then paste the Worker URL into **Interactions Endpoint URL** in the Discord
Developer Portal. Discord sends a signed PING to that URL and only saves it if
the signature verifies.
