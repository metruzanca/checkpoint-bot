/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/metruzanca/checkpoint-bot/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// clearCmd represents the clear command
var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all commands",

	Run: func(cmd *cobra.Command, args []string) {
		token := viper.GetString("TOKEN")
		dbPath := viper.GetString("DB_PATH")

		bot := server.NewBot(token, dbPath)
		bot.Start()
		defer bot.Stop()
		bot.UnregisterCommands()
	},
}

func init() {
	commandCmd.AddCommand(clearCmd)
}
