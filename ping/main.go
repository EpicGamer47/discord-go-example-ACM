package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/lus/dgc"
)

// slash commands
var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Responds with Pong!",
	},
}
var slashCommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Pong!",
			},
		})
	},
}

// text commands
func registerCommands(router *dgc.Router) {

	router.RegisterCmd(&dgc.Command{
		Name:        "ping",
		Description: "Responds with pong",
		Handler:     pingHandler,
		IgnoreCase:  true,
	})
}

func pingHandler(ctx *dgc.Ctx) {
	ctx.RespondText("Pong!")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	Token := os.Getenv("DISCORD_BOT_TOKEN")
	// NOTE: Changed "DISCORD_TOKEN" to "DISCORD_BOT_TOKEN" to match common practice
	if Token == "" {
		log.Fatal("Error: DISCORD_BOT_TOKEN not found in .env file.")
	}

	guildID := os.Getenv("TEST_GUILD_ID")
	if guildID == "" {
		log.Println("WARNING: TEST_GUILD_ID not found in .env. Registering commands globally (takes up to 1 hour).")
	} else {
		log.Printf("INFO: Registering commands to Guild ID: %s (instant update)", guildID)
	}

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	router := dgc.Create(&dgc.Router{
		Prefixes: []string{"!"},
	})

	registerCommands(router)

	dg.AddHandler(router.Handler())

	dg.Identify.Intents = (discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsMessageContent)

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			if handler, ok := slashCommandHandlers[i.ApplicationCommandData().Name]; ok {
				handler(s, i)
			}
		}
	})

	registeredCommands := make(map[string]string, len(commands))

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	log.Println("Bot is now running. Registering slash commands...")

	for _, v := range commands {
		cmd, err := dg.ApplicationCommandCreate(dg.State.User.ID, guildID, v)
		if err != nil {
			log.Fatalf("Cannot create command %s: %v", v.Name, err)
		}
		registeredCommands[cmd.Name] = cmd.ID
	}

	fmt.Println("Commands registered.")
	fmt.Println("Press CTRL-C to exit and automatically clean up slash commands.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("Shutting down gracefully...")
	for name, id := range registeredCommands {
		err := dg.ApplicationCommandDelete(dg.State.User.ID, guildID, id)
		if err != nil {
			log.Printf("Cannot delete command %s: %v", name, err)
		}
	}

	dg.Close()
	log.Println("Shutdown complete.")
}
