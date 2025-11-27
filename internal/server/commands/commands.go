package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/metruzanca/checkpoint-bot/internal/database"
)

// Global mutable map, only modified in init() functions
var commands = make(map[string]*Command)

type CommandHandler struct {
	DiscordClient *discordgo.Session
	Database      database.CheckpointDatabase
}

func NewCommandHandler(discordClient *discordgo.Session, database database.CheckpointDatabase) *CommandHandler {
	clearUnregisteredCommands(discordClient)
	return &CommandHandler{
		Database:      database,
		DiscordClient: discordClient,
	}
}

type Command struct {
	discordgo.ApplicationCommand
	Handler func(db database.CheckpointDatabase, s *discordgo.Session, i *discordgo.InteractionCreate)
}

func (h *CommandHandler) RegisterCommands() {
	for _, guild := range h.DiscordClient.State.Guilds {
		for _, cmd := range commands {
			registeredCmd, err := h.DiscordClient.ApplicationCommandCreate(h.DiscordClient.State.User.ID, guild.ID, &cmd.ApplicationCommand)
			if err != nil {
				log.Error("cannot register new command", "command", cmd.Name, "guild", guild.ID, "err", err)
				continue
			}
			log.Debug("Registered new command", "command", registeredCmd.Name, "guild", guild.ID)
		}

	}
	h.DiscordClient.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if cmd, ok := commands[i.ApplicationCommandData().Name]; ok {
			cmd.Handler(h.Database, s, i)
		}
	})
}

func registerCommand(cmd *Command) {
	commands[cmd.ApplicationCommand.Name] = cmd
}

func clearUnregisteredCommands(discordClient *discordgo.Session) {
	for _, guild := range discordClient.State.Guilds {
		registeredCommands, err := discordClient.ApplicationCommands(discordClient.State.User.ID, guild.ID)
		if err != nil {
			log.Error("Failed to get registered commands", "guild", guild.ID, "err", err)
			continue
		}
		for _, cmd := range registeredCommands {
			// Only delete commands that are not in the commands map
			if _, exists := commands[cmd.Name]; !exists {
				err := discordClient.ApplicationCommandDelete(discordClient.State.User.ID, guild.ID, cmd.ID)
				if err != nil {
					log.Error("Failed to delete command", "command", cmd.Name, "guild", guild.ID, "err", err)
				} else {
					log.Debug("Deleted unregistered command", "command", cmd.Name, "guild", guild.ID)
				}
			}
		}
	}
}

func ErrorResponse(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
}
