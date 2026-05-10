package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
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

	schema := `
	CREATE TABLE IF NOT EXISTS peliculas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nombre TEXT NOT NULL,
		agregada_por TEXT NOT NULL,
		creada_en DATETIME DEFAULT CURRENT_TIMESTAMP,
		vista_en DATETIME
	);`

	if _, err := db.Exec(schema); err != nil {
		log.Fatal(err)
	}
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
}

var handlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"agregar": handleAdd,
	"sugerir": handleSuggest,
	"vista":   handleMarkMovieAsSeen,
	"lista":   handleGetMovieList,
}


func handleAdd(s *discordgo.Session, i *discordgo.InteractionCreate) {
	nombre := i.ApplicationCommandData().Options[0].StringValue()
	userID := i.Member.User.ID

	res, err := db.Exec(
		`INSERT INTO peliculas (nombre, agregada_por) VALUES (?, ?)`,
		nombre, userID,
	)
	if err != nil {
		respond(s, i, "Error al agregar la película: "+err.Error())
		return
	}
	id, _ := res.LastInsertId()
	respond(s, i, fmt.Sprintf("✅ Agregada: **%s** (id %d)", nombre, id))
}

func handleSuggest(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var nombre string
	err := db.QueryRow(`
		SELECT nombre FROM peliculas
		WHERE vista_en IS NULL
		ORDER BY RANDOM()
		LIMIT 1
	`).Scan(&nombre)

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
	data := i.ApplicationCommandData()

	// Autocomplete
	if i.Type == discordgo.InteractionApplicationCommandAutocomplete {
		var query string
		for _, opt := range data.Options {
			if opt.Focused {
				query = opt.StringValue()
			}
		}
		respondAutocomplete(s, i, i.Member.User.ID, query)
		return
	}

	nombre := data.Options[0].StringValue()

	var peliID int
	err := db.QueryRow(`SELECT id FROM peliculas WHERE nombre = ? COLLATE NOCASE`, nombre).Scan(&peliID)
	if err == sql.ErrNoRows {
		respond(s, i, fmt.Sprintf("No encontré **%s** en el catálogo.", nombre))
		return
	}
	if err != nil {
		respond(s, i, "Error: "+err.Error())
		return
	}
	_, err = db.Exec(`UPDATE peliculas SET vista_en = CURRENT_TIMESTAMP WHERE id = ?`, peliID)
	if err != nil {
		respond(s, i, "Error al marcar como vista: "+err.Error())
		return
	}
	respond(s, i, fmt.Sprintf("👁️ Marcaste **%s** como vista.", nombre))
}

func handleGetMovieList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	rows, err := db.Query(`
		SELECT nombre,
			   CASE WHEN vista_en IS NOT NULL THEN 1 ELSE 0 END AS vista
		FROM peliculas
		LIMIT 25;
	`) // It'd be interesting add pagination to this feature

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

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: content},
	})
}

func respondAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate, userID, query string) {
	rows, err := db.Query(`
		SELECT nombre FROM peliculas
		WHERE nombre LIKE ? COLLATE NOCASE
			AND vista_en IS NULL
	`, "%"+query+"%")
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
