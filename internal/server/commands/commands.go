package commands

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/metruzanca/checkpoint-bot/internal/database"
)

// Global mutable map, only modified in init() functions
var commands = make(map[string]*Command)

// CommandHandler manages Discord slash command registration and execution
type CommandHandler struct {
	DiscordClient *discordgo.Session
	Database      database.CheckpointDatabase
}

// NewCommandHandler creates a new command handler and clears unregistered commands
func NewCommandHandler(discordClient *discordgo.Session, database database.CheckpointDatabase) *CommandHandler {
	clearUnregisteredCommands(discordClient)
	return &CommandHandler{
		Database:      database,
		DiscordClient: discordClient,
	}
}

// Command represents a Discord slash command with its handler function
type Command struct {
	discordgo.ApplicationCommand
	Handler func(db database.CheckpointDatabase, s *discordgo.Session, i *discordgo.InteractionCreate)
}

// RegisterCommands registers all commands for all guilds and sets up interaction handlers
func (h *CommandHandler) RegisterCommands() {
	for _, guild := range h.DiscordClient.State.Guilds {
		h.RegisterCommandsForGuild(guild.ID)
	}
	h.DiscordClient.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Handle modal submissions
		if i.Type == discordgo.InteractionModalSubmit {
			data := i.ModalSubmitData()
			log.Info("modal submitted", "custom_id", data.CustomID)
			if len(data.CustomID) > 0 && data.CustomID[:10] == "goal_modal_" {
				HandleGoalModalSubmission(h.Database, s, i)
				return
			}
		}

		// Handle application commands
		if i.Type == discordgo.InteractionApplicationCommand {
			commandName := i.ApplicationCommandData().Name
			userID := ""
			if i.Member != nil {
				userID = i.Member.User.ID
			} else if i.User != nil {
				userID = i.User.ID
			}

			// Rate limiting: check if user has exceeded rate limit
			if !commandRateLimiter.allow(userID) {
				log.Warn("rate limit exceeded", "command", commandName, "user", userID, "channel", i.ChannelID, "guild", i.GuildID)
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "You're using commands too quickly. Please wait a moment and try again.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}

			if cmd, ok := commands[commandName]; ok {
				log.Info("command executed", "command", commandName, "channel", i.ChannelID, "guild", i.GuildID, "user", userID)
				cmd.Handler(h.Database, s, i)
			} else {
				log.Warn("unknown command received", "command", commandName, "channel", i.ChannelID, "guild", i.GuildID)
			}
		}
	})
}

// RegisterCommandsForGuild registers all commands for a specific guild
func (h *CommandHandler) RegisterCommandsForGuild(guildID string) {
	for _, cmd := range commands {
		registeredCmd, err := h.DiscordClient.ApplicationCommandCreate(h.DiscordClient.State.User.ID, guildID, &cmd.ApplicationCommand)
		if err != nil {
			log.Error("cannot register new command", "command", cmd.Name, "guild", guildID, "err", err)
			continue
		}
		log.Debug("Registered new command", "command", registeredCmd.Name, "guild", guildID)
	}
}

// registerCommand registers a command in the global commands map
// This is called from init() functions in command files
func registerCommand(cmd *Command) {
	commands[cmd.ApplicationCommand.Name] = cmd
}

// GetAvailableCommands returns a list of all registered command names and descriptions
func GetAvailableCommands() []struct {
	Name        string
	Description string
} {
	result := make([]struct {
		Name        string
		Description string
	}, 0, len(commands))
	for _, cmd := range commands {
		result = append(result, struct {
			Name        string
			Description string
		}{
			Name:        cmd.Name,
			Description: cmd.Description,
		})
	}
	return result
}

// clearUnregisteredCommands removes commands from Discord that are no longer in the commands map
// This is called during command handler initialization to clean up old commands
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

// dbContext creates a context with timeout for database operations.
// Returns a context that will be cancelled after the timeout duration.
// The caller should defer cancel() to ensure proper cleanup.
func dbContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
