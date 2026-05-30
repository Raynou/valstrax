package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	_ "modernc.org/sqlite"
)

var (
	db      *sql.DB
	guildID = os.Getenv("GUILD_ID") // opcional: registro instantáneo en un solo server
)

func initDB() {
	var err error
	db, err = sql.Open("sqlite", "/tmp/peliculas.db")
	if err != nil {
		log.Fatal(err)
	}

	// Task 1.3: schema con guild_id para bases de datos nuevas
	schema := `
	CREATE TABLE IF NOT EXISTS peliculas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nombre TEXT NOT NULL,
		agregada_por TEXT NOT NULL,
		guild_id TEXT NOT NULL DEFAULT '',
		creada_en DATETIME DEFAULT CURRENT_TIMESTAMP,
		vista_en DATETIME
	);`

	if _, err := db.Exec(schema); err != nil {
		log.Fatal(err)
	}

	// Task 1.1: migración idempotente para bases de datos existentes
	_, err = db.Exec(`ALTER TABLE peliculas ADD COLUMN guild_id TEXT NOT NULL DEFAULT ''`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") && !strings.Contains(err.Error(), "already has column") {
		log.Fatal(err)
	}

	// Task 1.2: índice para optimizar filtros por guild
	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_peliculas_guild ON peliculas(guild_id)`); err != nil {
		log.Fatal(err)
	}
}

// Task 2.1: helper que extrae el guild_id de una interacción
func interactionGuildID(i *discordgo.InteractionCreate) (string, bool) {
	if i.GuildID == "" {
		return "", false
	}
	return i.GuildID, true
}

// Slash commands
var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "agregar",
		Description: "Agrega una película al catálogo",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "nombre",
				Description: "Nombre de la película",
				Required:    true,
			},
		},
	},
	{
		Name:        "sugerir",
		Description: "Sugiere una película que aún no hayas visto",
	},
	{
		Name:        "vista",
		Description: "Marca una película como vista por ti",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:         discordgo.ApplicationCommandOptionString,
				Name:         "nombre",
				Description:  "Película a marcar como vista",
				Required:     true,
				Autocomplete: true,
			},
		},
	},
	{
		Name:        "lista",
		Description: "Muestra el catálogo de películas",
	},
	{
		Name:        "quitar",
		Description: "Elimina una pelicula de la lista de pendientes",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:         discordgo.ApplicationCommandOptionString,
				Name:         "nombre",
				Description:  "Nombre de la película a eliminar",
				Required:     false,
				Autocomplete: true,
			},
			{
				Type:         discordgo.ApplicationCommandOptionInteger,
				Name:         "id",
				Description:  "ID de la película a eliminar",
				Required:     false,
				Autocomplete: false,
			},
		},
	},
	{
		Name:        "desmarcar",
		Description: "Marca una pelicula como no vista",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:         discordgo.ApplicationCommandOptionString,
				Name:         "nombre",
				Description:  "Nombre de la pelicula",
				Required:     false,
				Autocomplete: true,
			},
			{
				Type:         discordgo.ApplicationCommandOptionInteger,
				Name:         "id",
				Description:  "Id de la pelicula",
				Required:     false,
				Autocomplete: false,
			},
		},
	},
}

var handlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"agregar":   handleAdd,
	"sugerir":   handleSuggest,
	"vista":     handleMarkMovieAsSeen,
	"lista":     handleGetMovieList,
	"quitar":    handleRemove,
	"desmarcar": handleUnmarkMovie,
}

func handleAdd(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Task 2.2
	gID, ok := interactionGuildID(i)
	if !ok {
		respond(s, i, "Este comando solo está disponible dentro de un servidor.")
		return
	}

	name := i.ApplicationCommandData().Options[0].StringValue()
	userID := i.Member.User.ID

	// Task 3.1: incluir guild_id en el INSERT
	res, err := db.Exec(
		`INSERT INTO peliculas (nombre, agregada_por, guild_id) VALUES (?, ?, ?)`,
		name,
		userID,
		gID,
	)
	if err != nil {
		respond(s, i, "Error al agregar la película: "+err.Error())
		return
	}
	id, _ := res.LastInsertId()
	respond(s, i, fmt.Sprintf("✅ Agregada: **%s** (id %d)", name, id))
}

func handleRemove(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Task 2.2
	gID, ok := interactionGuildID(i)
	if !ok {
		respond(s, i, "Este comando solo está disponible dentro de un servidor.")
		return
	}

	data := i.ApplicationCommandData()

	// Autocomplete — Task 5.3: pasar i.GuildID
	if i.Type == discordgo.InteractionApplicationCommandAutocomplete {
		var query string
		for _, opt := range data.Options {
			if opt.Focused && opt.Name == "nombre" {
				query = opt.StringValue()
			}
		}
		respondAutocomplete(s, i, gID, query)
		return
	}
	var (
		name string
		id   int64
	)

	for _, opt := range data.Options {
		switch opt.Name {
		case "nombre":
			name = opt.StringValue()
		case "id":
			id = opt.IntValue()
		}
	}

	if name == "" && id == 0 {
		respond(s, i, "Debes proporcionar `nombre` o `id`.")
		return
	}

	// Task 3.2: filtrar DELETE por guild_id
	res, err := db.Exec(
		`DELETE FROM peliculas WHERE guild_id = ? AND ((? <> '' AND nombre = ? COLLATE NOCASE) OR (? <> 0 AND id = ?))`,
		gID,
		name,
		name,
		id,
		id,
	)

	if err != nil {
		respond(s, i, "Error al eliminar pelicula: "+err.Error())
		return
	}

	n, _ := res.RowsAffected()

	if n == 0 {
		respond(s, i, "No encontre ninguna pelicula con esos datos")
		return
	}

	respond(s, i, fmt.Sprintf("📤 Eliminada(s) %d película(s).", n))
}

func handleSuggest(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Task 2.2
	gID, ok := interactionGuildID(i)
	if !ok {
		respond(s, i, "Este comando solo está disponible dentro de un servidor.")
		return
	}

	var nombre string
	// Task 4.1: filtrar por guild_id
	err := db.QueryRow(`
		SELECT nombre FROM peliculas
		WHERE vista_en IS NULL
			AND guild_id = ?
		ORDER BY RANDOM()
		LIMIT 1
	`, gID).Scan(&nombre)

	if err == sql.ErrNoRows {
		respond(s, i, "🎬 ¡Ya viste todo lo que hay en el catálogo! Agrega más con `/agregar`.")
		return
	}
	if err != nil {
		respond(s, i, "Error al buscar sugerencias: "+err.Error())
		return
	}

	respond(s, i, fmt.Sprintf("🎲 Sugerencia: **%s**\nMárcala como vista con `/vista` cuando termines.", nombre))
}

func handleMarkMovieAsSeen(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Task 2.2
	gID, ok := interactionGuildID(i)
	if !ok {
		respond(s, i, "Este comando solo está disponible dentro de un servidor.")
		return
	}

	data := i.ApplicationCommandData()

	// Autocomplete — Task 5.3: pasar i.GuildID
	if i.Type == discordgo.InteractionApplicationCommandAutocomplete {
		var query string
		for _, opt := range data.Options {
			if opt.Focused {
				query = opt.StringValue()
			}
		}
		respondAutocomplete(s, i, gID, query)
		return
	}

	nombre := data.Options[0].StringValue()

	// Task 3.3: filtrar SELECT y UPDATE por guild_id
	var peliID int
	err := db.QueryRow(`SELECT id FROM peliculas WHERE nombre = ? COLLATE NOCASE AND guild_id = ?`, nombre, gID).Scan(&peliID)
	if err == sql.ErrNoRows {
		respond(s, i, fmt.Sprintf("No encontré **%s** en el catálogo.", nombre))
		return
	}
	if err != nil {
		respond(s, i, "Error: "+err.Error())
		return
	}
	_, err = db.Exec(`UPDATE peliculas SET vista_en = CURRENT_TIMESTAMP WHERE id = ? AND guild_id = ?`, peliID, gID)
	if err != nil {
		respond(s, i, "Error al marcar como vista: "+err.Error())
		return
	}
	respond(s, i, fmt.Sprintf("👁️ Marcaste **%s** como vista.", nombre))
}

func handleGetMovieList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Task 2.2
	gID, ok := interactionGuildID(i)
	if !ok {
		respond(s, i, "Este comando solo está disponible dentro de un servidor.")
		return
	}

	// Task 4.2: filtrar por guild_id
	rows, err := db.Query(`
		SELECT nombre,
			   CASE WHEN vista_en IS NOT NULL THEN 1 ELSE 0 END AS vista
		FROM peliculas
		WHERE guild_id = ?
		LIMIT 25;
	`, gID) // It'd be interesting add pagination to this feature

	if err != nil {
		respond(s, i, "Error: "+err.Error())
		return
	}
	defer rows.Close()

	msg := "**Catálogo:**\n"
	count := 0
	for rows.Next() {
		var nombre string
		var vista int
		rows.Scan(&nombre, &vista)
		marca := "⬜"
		if vista == 1 {
			marca = "✅"
		}
		msg += fmt.Sprintf("%s %s\n", marca, nombre)
		count++
	}
	if count == 0 {
		msg = "El catálogo está vacío. Agrega algo con `/agregar`."
	}
	respond(s, i, msg)
}

func handleUnmarkMovie(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Task 2.2
	gID, ok := interactionGuildID(i)
	if !ok {
		respond(s, i, "Este comando solo está disponible dentro de un servidor.")
		return
	}

	data := i.ApplicationCommandData()

	// Autocomplete — Task 5.3: pasar i.GuildID
	if i.Type == discordgo.InteractionApplicationCommandAutocomplete {
		var query string
		for _, opt := range data.Options {
			if opt.Focused {
				query = opt.StringValue()
			}
		}
		respondAutocompleteSeenMovies(s, i, gID, query)
		return
	}

	var (
		name string
		id   int64
	)

	for _, opt := range data.Options {
		switch opt.Name {
		case "nombre":
			name = opt.StringValue()
		case "id":
			id = opt.IntValue()
		}
	}

	// Task 3.4: filtrar UPDATEs por guild_id
	if name != "" {
		res, err := db.Exec(
			`UPDATE peliculas SET vista_en = NULL WHERE nombre = ? AND guild_id = ?`,
			name,
			gID,
		)

		if err != nil {
			respond(s, i, fmt.Sprintf("Error al momento de desmarcar pelicula: %s", name))
		}

		n, _ := res.RowsAffected()

		if n == 0 {
			respond(s, i, fmt.Sprintf("No encontre ninguna pelicula con el nombre: %s", name))
		}

		respond(s, i, fmt.Sprintf("Pelicula %s marcada como no vista", name))

	} else if id != 0 {
		res, err := db.Exec(
			`UPDATE peliculas SET vista_en = NULL WHERE id = ? AND guild_id = ?`,
			id,
			gID,
		)

		if err != nil {
			respond(s, i, fmt.Sprintf("Error al momento de desmarcar pelicula con id: `%d`", id))
		}

		n, _ := res.RowsAffected()

		if n == 0 {
			respond(s, i, fmt.Sprintf("No encontré ninguna pelicula con la id: `%d`", id))
		}

		respond(s, i, fmt.Sprintf("Pelicula con id `%d` marcada como no vista", id))
	}
}

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: content},
	})
}

// Task 5.1: recibe guildID y filtra el SELECT por él
func respondAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate, guildID, query string) {
	rows, err := db.Query(`
		SELECT nombre FROM peliculas
		WHERE nombre LIKE ? COLLATE NOCASE
			AND vista_en IS NULL
			AND guild_id = ?
	`, "%"+query+"%", guildID)
	if err != nil {
		return
	}
	defer rows.Close()

	var choices []*discordgo.ApplicationCommandOptionChoice
	for rows.Next() {
		var n string
		rows.Scan(&n)
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name: n, Value: n,
		})
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{Choices: choices},
	})
}

// Task 5.2: recibe guildID y filtra el SELECT por él
// TODO: I don't like this, figure out how to don't repeat the logic from the above method
func respondAutocompleteSeenMovies(s *discordgo.Session, i *discordgo.InteractionCreate, guildID, query string) {
	rows, err := db.Query(
		`SELECT nombre FROM peliculas
		WHERE nombre LIKE ? COLLATE NOCASE
			AND vista_en IS NOT NULL
			AND guild_id = ?
		`, "%"+query+"%", guildID,
	)

	if err != nil {
		return
	}

	defer rows.Close()

	var choices []*discordgo.ApplicationCommandOptionChoice
	for rows.Next() {
		var n string
		rows.Scan(&n)
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name: n, Value: n,
		})
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{Choices: choices},
	})
}

func main() {
	initDB()
	defer db.Close()

	dg, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		name := i.ApplicationCommandData().Name
		if h, ok := handlers[name]; ok {
			h(s, i)
		}
	})

	if err := dg.Open(); err != nil {
		log.Fatal(err)
	}
	defer dg.Close()

	// Registrar comandos
	registered, err := dg.ApplicationCommandBulkOverwrite(dg.State.User.ID, guildID, commands)
	if err != nil {
		log.Fatal("Error registrando comandos: ", err)
	}
	log.Printf("Bot listo. %d comandos registrados.", len(registered))

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc
}
