package cmd

import (
	"fmt"

	"github.com/charmbracelet/log"

	"github.com/metruzanca/checkpoint-bot/internal/config"
	"github.com/metruzanca/checkpoint-bot/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// inviteCmd represents the invite command
var inviteCmd = &cobra.Command{
	Use:               "invite",
	Short:             "Generates a Discord invite link for the bot",
	PersistentPreRunE: config.PersistentPreRunE,
	Run: func(cmd *cobra.Command, args []string) {
		clientID := viper.GetString("CLIENT_ID")
		if clientID == "" {
			log.Fatal("CLIENT_ID is not set")
		}

		inviteLink := util.GetInviteLink(clientID)
		fmt.Println("Invite link: ", inviteLink)
	},
}

func init() {
	rootCmd.AddCommand(inviteCmd)
}
