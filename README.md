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

Setting up the project requires a `DISCORD_TOKEN` and a `GUILD_ID`.

Run this command for starting the bot:

```bash
DISCORD_TOKEN=<your_token> GUILD_ID=<your_guild_id> go main.go
```

