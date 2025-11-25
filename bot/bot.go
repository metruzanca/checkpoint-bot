package bot

import (
	"github.com/charmbracelet/log"
	"github.com/metruzanca/checkpoint-bot/internal/database"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	Session *discordgo.Session
	token   string
	db      database.CheckpointDatabase
}

func NewBot(token string) *Bot {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("Error creating Discord session: ", "err", err)
	}
	return &Bot{
		Session: session,
		token:   token,
		db:      database.NewSQLiteCheckpointDatabase(),
	}
}

func (b *Bot) Start() {
	err := b.Session.Open()
	if err != nil {
		log.Fatal("Error opening Discord session: ", "err", err)
	}
	defer b.Session.Close()
	log.Info("Bot is now running.")
}

func (b *Bot) Stop() {
	b.Session.Close()
	log.Info("Bot stopped gracefully.")
}

// func (b *Bot) SendMessage(channelID, message string) {
// 	channel, err := b.Session.Channel(channelID)
// 	if err != nil {
// 		log.Fatal("Error getting channel: ", "err", err)
// 	}

// 	_, err = b.Session.ChannelMessageSend(channel.ID, message)
// 	if err != nil {
// 		log.Fatal("Error sending message: ", "err", err)
// 	}
// }

func (b *Bot) RegisterCommands() {

}
