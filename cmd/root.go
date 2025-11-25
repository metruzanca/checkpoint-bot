package cmd

import (
	"os"

	"github.com/metruzanca/checkpoint-bot/bot"
	"github.com/metruzanca/checkpoint-bot/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:                "checkpoint",
	Short:              "A ",
	PersistentPostRunE: config.PersistentPostRunE,
	Run: func(cmd *cobra.Command, args []string) {
		config.LoadConfig()
		token := viper.GetString("DISCORD_CLIENT_TOKEN")
		channelID := viper.GetString("DISCORD_CHANNEL_ID")

		bot := bot.NewBot(token)
		bot.Start()
		bot.SendMessage(channelID, "Bot is now running.")

		<-make(chan struct{})
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("token", "", "Discord bot token")
	rootCmd.PersistentFlags().String("channel-id", "", "Discord channel ID")
}
