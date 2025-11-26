package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/metruzanca/checkpoint-bot/internal/database"
	"github.com/metruzanca/checkpoint-bot/internal/server/commands"
	"github.com/metruzanca/checkpoint-bot/internal/sqlc"
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

func (b *Bot) RegisterCommand(guildID string, cmd *discordgo.ApplicationCommand, callback func(s *discordgo.Session, i *discordgo.InteractionCreate)) error {
	registeredCmd, err := b.DiscordClient.ApplicationCommandCreate(b.DiscordClient.State.User.ID, guildID, cmd)
	if err != nil {
		return fmt.Errorf("cannot create command %s in guild %s: %w", cmd.Name, guildID, err)
	}

	b.DiscordClient.AddHandler(callback)
	log.Debug("Registered command", "command", registeredCmd.Name, "guild", guildID)

	return nil
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
		err := b.RegisterCommand(guild.ID, commands.CreateCheckpoint, func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			checkpoint, err := b.Database.CreateCheckpoint(context.Background(), sqlc.CreateCheckpointParams{
				ScheduledAt: time.Now().Add(time.Hour * 24 * 14).Format(time.RFC3339),
				ChannelID:   i.ChannelID,
			})

			if err != nil {
				log.Error("cannot create checkpoint", "err", err)
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Error creating checkpoint",
					},
				})
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Checkpoint created: %s", checkpoint.ID),
				},
			})
		})
		// If a command fails to register, log the error and continue
		if err != nil {
			log.Error("cannot register command", "command", commands.CreateCheckpoint.Name, "guild", guild.ID, "err", err)
		}
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

func (b *Bot) Stop() {
	// Unregister commands that we've registered from all guilds
	// alternatively, delete all commands with:
	// for _, guild := range b.Session.State.Guilds {
	// 	registeredCommands, err := b.Session.ApplicationCommands(b.Session.State.User.ID, guild.ID)
	// 	if err != nil {
	// 		log.Error("Failed to get registered commands", "guild", guild.ID, "err", err)
	// 		continue
	// 	}
	// 	for name, cmd := range registeredCommands {
	// 		err := b.Session.ApplicationCommandDelete(b.Session.State.User.ID, guild.ID, cmd.ID)
	// 		if err != nil {
	// 			log.Error("Failed to delete command", "command", name, "guild", guild.ID, "err", err)
	// 		}
	// 	}
	// }

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
