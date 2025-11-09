package main

import (
	"fmt"
	"math/rand"

	"github.com/bwmarrin/discordgo"
	"github.com/lus/dgc"
)

// =================================================================================
// COMPONENT DEFINITIONS
// =================================================================================

const (
	rollButtonID = "roll_dice_button_v1"
)

var ComponentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	rollButtonID: buttonRollHandler,
}

func buttonRollHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// get og author id
	authorID := i.Message.Interaction.User.ID
	// get button presser id
	presserID := i.Member.User.ID

	if authorID != presserID {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Unauthorized! Start your own roll.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	// Roll the dice, 1-6 inclusive
	roll := rand.Intn(6) + 1

	// Edit message
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("<@!%s> rolled a **%d**!", presserID, roll),
			// Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// =================================================================================
// SLASH COMMAND DEFINITIONS & HANDLERS
// =================================================================================

// discord frontend
var SlashCommands = []*discordgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Responds with Pong!",
	},
	{
		Name:        "avatar",
		Description: "Gets the avatar URL for a specified user.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "The user whose avatar you want to view.",
				Required:    false,
			},
		},
	},
	{
		Name:        "roll",
		Description: "Sends a message with a button to roll a 6-sided die.",
	},
}

// maps all commands to their callbacks
var SlashCommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"ping":   slashPingHandler,
	"avatar": slashAvatarHandler,
	"roll":   slashRollHandler,
}

// slash command callbacks
func slashPingHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong!",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func slashAvatarHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	// user to use...
	var targetUser *discordgo.User

	// options is how you get params
	options := data.Options

	// 1. Check if the 'user' option was provided (by checking if the options slice is non-empty)
	if len(options) > 0 && options[0].Name == "user" {
		// get the value of the first param as a string
		// discord sends ided stuff (users, roles, channels etc. )
		// as just their id
		userID := options[0].Value.(string)

		// so that you can look it up in the "Resolved" cache
		// TODO you can extract this resolve logic to a "getter" method
		user, found := data.Resolved.Users[userID]

		// this realistically should never error
		if !found {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Error: Could not find the specified user in the interaction data.",
				},
			})
			return
		}
		targetUser = user
	} else {
		// if no user provided, just use author
		// you can just get it off the interaction

		if i.Member != nil { // if ran in server
			targetUser = i.Member.User
		} else { // if ran in DMs
			targetUser = i.User
		}
	}

	// gets the avatar with size 2048px
	avatarURL := targetUser.AvatarURL("2048")

	// format & reply!
	content := fmt.Sprintf("**%s's Avatar URL:**\n%s", targetUser.Username, avatarURL)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// Sends the message with view & button immediately
func slashRollHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Click the button to roll a 6-sided die!",
			// Flags:   discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{ // must store buttons in rows
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label: "Roll",
							Style: discordgo.SuccessButton, // green
							Emoji: &discordgo.ComponentEmoji{
								Name: "🎲",
							},
							CustomID: rollButtonID,
						},
					},
				},
			},
		},
	})
}

// =================================================================================
// PREFIX COMMAND DEFINITIONS & HANDLERS
// =================================================================================

// registers all commands
// & gives info for help
func SetupPrefixCommands(router *dgc.Router) {
	router.RegisterCmd(&dgc.Command{
		Name:        "ping",
		Description: "Pongs the ping",
		Handler:     pingHandler,
		IgnoreCase:  true,
	})

	router.RegisterCmd(&dgc.Command{
		Name:        "avatar",
		Description: "Gets the avatar URL for a mentioned user or yourself.",
		Handler:     avatarHandler,
		IgnoreCase:  true,
	})

	router.RegisterCmd(&dgc.Command{
		Name:        "help",
		Description: "Sends a help message",
		Handler:     helpHandler,
		IgnoreCase:  true,
	})
}

// text callbacks
func pingHandler(ctx *dgc.Ctx) {
	ctx.RespondText("Pong!")
}

func avatarHandler(ctx *dgc.Ctx) {
	// just to be clear, ctx.Event corresponds
	// to the messageCreate event (the message)

	// While we can't check if the command has a user,
	// we can find the mentioned users
	// and get the first one

	var targetUser *discordgo.User
	if len(ctx.Event.Mentions) > 0 {
		targetUser = ctx.Event.Mentions[0]
	} else {
		// default to author
		targetUser = ctx.Event.Author
	}

	// gets the avatar with size 2048px
	avatarURL := targetUser.AvatarURL("2048")

	// format and send
	response := fmt.Sprintf("**%s's Avatar URL:**\n%s", targetUser.Username, avatarURL)

	ctx.RespondText(response)
}

// sends a simple help command
// just sends command and short description
// i dont feel like figuring out how to show params
func helpHandler(ctx *dgc.Ctx) {
	var response string

	response += "**Available Prefix Commands (Prefix: !):**\n"

	for _, cmd := range ctx.Router.Commands {
		if cmd.Name != "" {
			response += fmt.Sprintf("`!%s`: %s\n", cmd.Name, cmd.Description)
		}
	}

	ctx.RespondText(response)
}
