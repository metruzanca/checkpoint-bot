package bot

import (
	"github.com/charmbracelet/log"
	"github.com/metruzanca/checkpoint-bot/internal/database"
	"github.com/spf13/viper"

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

var commands = make(map[string]*discordgo.ApplicationCommand)
var commandCallbacks = make(map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate))

// RegisterCommand registers a new command with the bot, must be called before Start()
func (b *Bot) RegisterCommand(
	cmd *discordgo.ApplicationCommand,
	callback func(s *discordgo.Session, i *discordgo.InteractionCreate),
) {
	commands[cmd.Name] = cmd
	commandCallbacks[cmd.Name] = callback
}

func (b *Bot) Start() {
	b.Session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Info("Logged in as", "username", s.State.User.Username+s.State.User.Discriminator)
	})

	err := b.Session.Open()
	if err != nil {
		log.Fatal("Error opening Discord session: ", "err", err)
	}

	log.Debug("Registering commands", "commands", len(commands), "guilds", len(b.Session.State.Guilds))

	for _, guild := range b.Session.State.Guilds {
		for name, cmd := range commands {
			registeredCmd, err := b.Session.ApplicationCommandCreate(b.Session.State.User.ID, guild.ID, cmd)
			if err != nil {
				log.Fatal("Cannot create command", "command", name, "err", err)
			}

			b.Session.AddHandler(commandCallbacks[cmd.Name])
			log.Debug("Registered command", "command", registeredCmd.Name, "guild", guild.ID)
		}
	}

	// Send a message to the dev channel if it is set
	channelID := viper.GetString("CHANNEL_ID")
	if channelID != "" {
		b.SendTextMessage(channelID, "Bot has restarted.")
	}
}

func (b *Bot) Stop() {
	// Unregister commands that we've registered from all guilds
	// alternatively, delete all commands with:
	for _, guild := range b.Session.State.Guilds {
		registeredCommands, err := b.Session.ApplicationCommands(b.Session.State.User.ID, guild.ID)
		if err != nil {
			log.Error("Failed to get registered commands", "guild", guild.ID, "err", err)
			continue
		}
		for name, cmd := range registeredCommands {
			err := b.Session.ApplicationCommandDelete(b.Session.State.User.ID, guild.ID, cmd.ID)
			if err != nil {
				log.Error("Failed to delete command", "command", name, "guild", guild.ID, "err", err)
			}
		}
	}

	b.Session.Close()

	log.Info("Bot stopped gracefully.")
}

func (b *Bot) SendTextMessage(channelID, message string) {
	channel, err := b.Session.Channel(channelID)
	if err != nil {
		log.Fatal("Error getting channel: ", "err", err)
	}

	_, err = b.Session.ChannelMessageSend(channel.ID, message)
	if err != nil {
		log.Fatal("Error sending message: ", "err", err)
	}

	log.Info("Message sent to channel: ", "channelID", channel.ID, "message", message)
}
