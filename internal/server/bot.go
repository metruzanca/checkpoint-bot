package server

import (
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/metruzanca/checkpoint-bot/internal/database"
	"github.com/metruzanca/checkpoint-bot/internal/database/sqlite"
	"github.com/metruzanca/checkpoint-bot/internal/server/commands"
	"github.com/metruzanca/checkpoint-bot/internal/util"
	"github.com/spf13/viper"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	DiscordClient  *discordgo.Session
	Database       database.CheckpointDatabase
	CommandHandler *commands.CommandHandler
}

func NewBot(token string, dbPath string) *Bot {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("Error creating Discord session: ", "err", err)
	}
	return &Bot{
		DiscordClient: session,
		Database:      sqlite.NewSqliteDatabase(dbPath),
	}
}

var startTime time.Time

func init() {
	startTime = time.Now()
}

func (b *Bot) Start() error {
	// Startup logging
	b.DiscordClient.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Info("Logged in as", "username", s.State.User.Username+"#"+s.State.User.Discriminator, "bot_id", s.State.User.ID)
		fmt.Println("Invite me to a server: ", util.GetInviteLink(s.State.User.ID))

		// Show connected guilds
		guilds := b.DiscordClient.State.Guilds
		for _, guild := range guilds {
			if guild.Name == "" {
				// I think `State.Guilds` is a cache, sometimes it's empty. Weird
				b.DiscordClient.Guild(guild.ID)
			}

			log.Info("Guild", "id", guild.ID, "name", guild.Name, "member_count", guild.MemberCount)
		}

		// Show available commands
		registeredCommands := commands.GetAvailableCommands()
		for _, cmd := range registeredCommands {
			log.Info("Command", "name", cmd.Name, "description", cmd.Description)
		}

		// Send a message to the dev channel if it is set
		channelID := viper.GetString("CHANNEL_ID")
		if channelID != "" {
			b.SendTextMessage(channelID, "Bot has restarted.")
		}
	})

	err := b.DiscordClient.Open()
	if err != nil {
		return fmt.Errorf("Error opening Discord session: %w", err)
	}

	b.CommandHandler = commands.NewCommandHandler(b.DiscordClient, b.Database)
	b.CommandHandler.RegisterCommands()

	// Handle bot being added to a new server
	b.DiscordClient.AddHandler(b.onGuildJoined)

	return nil
}

func (b *Bot) onGuildJoined(s *discordgo.Session, g *discordgo.GuildCreate) {
	log.Info("Bot added to new guild", "guild_id", g.ID, "guild_name", g.Name, "member_count", g.MemberCount)
	b.CommandHandler.RegisterCommandsForGuild(g.ID)
}

func (b *Bot) UnregisterCommands() {
	for _, guild := range b.DiscordClient.State.Guilds {
		registeredCommands, err := b.DiscordClient.ApplicationCommands(b.DiscordClient.State.User.ID, guild.ID)
		if err != nil {
			log.Error("Failed to get registered commands", "guild", guild.ID, "err", err)
			continue
		}
		for _, cmd := range registeredCommands {
			b.DiscordClient.ApplicationCommandDelete(b.DiscordClient.State.User.ID, guild.ID, cmd.ID)
			if err != nil {
				log.Error("Failed to delete command", "command", cmd.Name, "guild", guild.ID, "err", err)
			}
		}
	}

	log.Info("Commands unregistered successfully.")
}

func (b *Bot) Stop() {
	b.DiscordClient.Close()
	b.Database.Close()

	log.Info("Bot stopped gracefully.", "runtime", time.Since(startTime))
}

func (b *Bot) SendTextMessage(channelID, message string) {
	_, err := b.DiscordClient.ChannelMessageSend(channelID, message)
	if err != nil {
		log.Error("Error sending message: ", "err", err)
	}

	log.Info("Message sent to channel: ", "channelID", channelID, "message", message)
}
