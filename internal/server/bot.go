package server

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/metruzanca/checkpoint-bot/internal/database"
	"github.com/metruzanca/checkpoint-bot/internal/server/commands"
	"github.com/metruzanca/checkpoint-bot/internal/server/handlers"
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

func (b *Bot) registerCommand(guildID string, cmd *discordgo.ApplicationCommand) {
	registeredCmd, err := b.DiscordClient.ApplicationCommandCreate(b.DiscordClient.State.User.ID, guildID, cmd)
	if err != nil {
		log.Error("cannot create command", "command", cmd.Name, "guild", guildID, "err", err)
		return
	}
	log.Debug("Registered command", "command", registeredCmd.Name, "guild", guildID)
}

func (b *Bot) registerHandlers() {
	checkpointHandler := handlers.NewCheckpointHandler(b.DiscordClient, b.Database)
	b.DiscordClient.AddHandler(checkpointHandler.CreateCheckpoint)
}

func (b *Bot) Start() error {

	b.DiscordClient.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Info("Logged in as", "username", s.State.User.Username+s.State.User.Discriminator)
	})

	err := b.DiscordClient.Open()
	if err != nil {
		return fmt.Errorf("Error opening Discord session: %w", err)
	}

	defer b.DiscordClient.Close()

	// Register commands for each guild

	for _, guild := range b.DiscordClient.State.Guilds {
		b.registerCommand(guild.ID, commands.CreateCheckpoint)
		b.registerHandlers()
	}

	// Send a message to the dev channel if it is set
	channelID := viper.GetString("CHANNEL_ID")
	if channelID != "" {
		b.SendTextMessage(channelID, "Bot has restarted.")
	}

	// Wait for interrupt signal to gracefully shutdown
	// Docker containers typically send SIGTERM when stopping
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	log.Info("Shutting down...")

	return nil
}

func (b *Bot) unregisterCommands() {
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
