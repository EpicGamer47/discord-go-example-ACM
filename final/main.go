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

// main
func main() {
	// loading variables & init errors
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	Token := os.Getenv("DISCORD_BOT_TOKEN")
	if Token == "" {
		log.Fatal("Error: DISCORD_BOT_TOKEN not found in .env file.")
	}

	// NOTE! If you use a test server, it will NOT register globally **AT ALL**
	guildID := os.Getenv("TEST_GUILD_ID")
	if guildID == "" {
		log.Println("WARNING: TEST_GUILD_ID not found in .env. Registering commands globally (takes up to 1 hour).")
	} else {
		log.Printf("INFO: Registering commands to Guild ID: %s (instant update)", guildID)
	}

	// starting discord handlers
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// go does not easily support dynamic prefixes...
	router := dgc.Create(&dgc.Router{
		Prefixes: []string{"!"},
	})

	// Use exported setup function from command_logic.go
	SetupPrefixCommands(router)

	dg.Identify.Intents = (discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsMessageContent)

	dg.AddHandler(router.Handler())

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		// slash commands
		case discordgo.InteractionApplicationCommand:
			if handler, ok := SlashCommandHandlers[i.ApplicationCommandData().Name]; ok {
				handler(s, i)
			}

		// button clicks
		case discordgo.InteractionMessageComponent:
			if handler, ok := ComponentHandlers[i.MessageComponentData().CustomID]; ok {
				handler(s, i)
			} else {
				log.Printf("Missing handler for button: %s", i.MessageComponentData().CustomID)
			}
		}
	})

	// Use exported slice length from command_logic.go
	registeredCommands := make(map[string]string, len(SlashCommands))

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	log.Println("Bot is now running. Registering slash commands...")

	// register slash commands
	// Iterate over exported slice from command_logic.go
	for _, v := range SlashCommands {
		cmd, err := dg.ApplicationCommandCreate(dg.State.User.ID, guildID, v)
		if err != nil {
			log.Fatalf("Cannot create command %s: %v", v.Name, err)
		}
		registeredCommands[cmd.Name] = cmd.ID
	}

	fmt.Println("Commands registered.")
	fmt.Println("Press CTRL-C to exit and automatically clean up slash commands.")

	// check for CTRL + C
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("Shutting down gracefully...")

	// you really shouldnt delete commands if you're registering globally lmao
	// useful for testing ig
	if guildID != "" {
		for name, id := range registeredCommands {
			err := dg.ApplicationCommandDelete(dg.State.User.ID, guildID, id)
			if err != nil {
				log.Printf("Cannot delete command %s: %v", name, err)
			}
		}
	}

	dg.Close()
	log.Println("Shutdown complete.")
}
