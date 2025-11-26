package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/metruzanca/checkpoint-bot/internal/database"
)

var commands = make(map[string]*Command)

type CommandHandler struct {
	DiscordClient *discordgo.Session
	Database      database.CheckpointDatabase
}

func NewCommandHandler(discordClient *discordgo.Session, database database.CheckpointDatabase) *CommandHandler {
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
				log.Error("cannot create command", "command", cmd.Name, "guild", guild.ID, "err", err)
				return
			}
			log.Debug("Registered command", "command", registeredCmd.Name, "guild", guild.ID)
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
