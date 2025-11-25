package cmd

import (
	"fmt"

	"github.com/charmbracelet/log"

	"github.com/metruzanca/checkpoint-bot/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// inviteCmd represents the invite command
var inviteCmd = &cobra.Command{
	Use:               "invite",
	Short:             "Generates a Discord invite link for the bot",
	PersistentPreRunE: config.PersistentPostRunE,
	Run: func(cmd *cobra.Command, args []string) {
		clientID := viper.GetString("DISCORD_CLIENT_ID")
		if clientID == "" {
			log.Fatal("DISCORD_CLIENT_ID is not set")
		}

		inviteLink := fmt.Sprintf("https://discord.com/api/oauth2/authorize?client_id=%s&permissions=8&scope=bot", clientID)
		fmt.Println("Invite link: ", inviteLink)
	},
}

func init() {
	rootCmd.AddCommand(inviteCmd)
}
