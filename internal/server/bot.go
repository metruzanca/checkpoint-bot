package server

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/metruzanca/checkpoint-bot/internal/database"
	"github.com/metruzanca/checkpoint-bot/internal/server/commands"
	"github.com/spf13/viper"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	DiscordClient *discordgo.Session
	Database      database.CheckpointDatabase
}

func NewBot(token string, dbPath string) *Bot {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("Error creating Discord session: ", "err", err)
	}
	return &Bot{
		DiscordClient: session,
		Database:      database.NewSQLiteCheckpointDatabase(dbPath),
	}
}

func (b *Bot) Start() error {

	b.DiscordClient.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Info("Logged in as", "username", s.State.User.Username+s.State.User.Discriminator)
	})

	err := b.DiscordClient.Open()
	if err != nil {
		return fmt.Errorf("Error opening Discord session: %w", err)
	}

	commandHandler := commands.NewCommandHandler(b.DiscordClient, b.Database)
	commandHandler.RegisterCommands()

	// Send a message to the dev channel if it is set
	channelID := viper.GetString("CHANNEL_ID")
	if channelID != "" {
		b.SendTextMessage(channelID, "Bot has restarted.")
	}

	return nil
}

func (b *Bot) UnregisterCommands() {
	for _, guild := range b.DiscordClient.State.Guilds {
		registeredCommands, err := b.DiscordClient.ApplicationCommands(b.DiscordClient.State.User.ID, guild.ID)
		if err != nil {
			log.Error("Failed to get registered commands", "guild", guild.ID, "err", err)
			continue
		}
		for name, cmd := range registeredCommands {
			err := b.DiscordClient.ApplicationCommandDelete(b.DiscordClient.State.User.ID, guild.ID, cmd.ID)
			if err != nil {
				log.Error("Failed to delete command", "command", name, "guild", guild.ID, "err", err)
			}
		}
	}
	log.Info("Commands unregistered successfully.")
}

func (b *Bot) Stop() {
	b.DiscordClient.Close()

	log.Info("Bot stopped gracefully.")
}

func (b *Bot) SendTextMessage(channelID, message string) {
	channel, err := b.DiscordClient.Channel(channelID)
	if err != nil {
		log.Fatal("Error getting channel: ", "err", err)
	}

	_, err = b.DiscordClient.ChannelMessageSend(channel.ID, message)
	if err != nil {
		log.Fatal("Error sending message: ", "err", err)
	}

	log.Info("Message sent to channel: ", "channelID", channel.ID, "message", message)
}
