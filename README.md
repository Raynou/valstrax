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
npm run migrate:local   # apply migrations/ to the local D1 database
npm run dev
```

Schema changes are versioned with Wrangler's D1 migrations. Add a new file to
`migrations/` (`wrangler d1 migrations create valstrax-db <name>`) and re-run
`npm run migrate:local`.

### Deployment

Every push to `main` triggers `.github/workflows/deploy-cloudflare.yml`, which
runs `typecheck` → `migrate:remote` → `deploy` → `register`. Nothing to do by
hand for a normal release.

**One-time setup.** The Worker's runtime secret stays in Cloudflare:

```bash
npx wrangler secret put DISCORD_PUBLIC_KEY
```

The pipeline authenticates through three **GitHub Actions secrets** (Settings →
Secrets and variables → Actions):

| Secret | Used by | How to get it |
| --- | --- | --- |
| `CLOUDFLARE_API_TOKEN` | `migrate:remote`, `deploy` | Cloudflare dashboard → My Profile → API Tokens → Create Token, permissions **Workers Scripts: Edit** and **D1: Edit** |
| `DISCORD_TOKEN` | `register` | Discord Developer Portal → your app → Bot → Token |
| `DISCORD_APPLICATION_ID` | `register` | Discord Developer Portal → your app → General Information → Application ID |

First deploy against the already-provisioned remote database: `migrate:remote`
creates the `d1_migrations` table and applies `0001` as a no-op (it is a copy of
the old `schema.sql`, all `IF NOT EXISTS`).

**Emergency manual deploy** (GitHub Actions unavailable): run the same scripts
locally with `CLOUDFLARE_API_TOKEN`, `DISCORD_TOKEN` and `DISCORD_APPLICATION_ID`
in your environment (or `.env`):

```bash
npm run ci   # typecheck && migrate:remote && deploy && register
```

After the first deploy, paste the Worker URL into **Interactions Endpoint URL**
in the Discord Developer Portal. Discord sends a signed PING to that URL and only
saves it if the signature verifies.
