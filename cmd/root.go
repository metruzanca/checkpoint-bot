package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/metruzanca/checkpoint-bot/internal/config"
	"github.com/metruzanca/checkpoint-bot/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:               "checkpoint",
	Short:             "A Discord bot for managing our accountability group's checkpoints",
	PersistentPreRunE: config.PersistentPreRunE,
	Run: func(cmd *cobra.Command, args []string) {
		token := viper.GetString("TOKEN")
		dbPath := viper.GetString("DB_PATH")

		bot := server.NewBot(token, dbPath)

		if err := bot.Start(); err != nil {
			log.Fatal("Error starting bot: %w", err)
		}

		// Wait for interrupt signal to gracefully shutdown
		// Docker containers typically send SIGTERM when stopping
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		<-sc
		log.Info("Shutting down...")
		bot.Stop()
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
	rootCmd.PersistentFlags().String("db-path", "", "Path to SQLite database file (default: checkpoint.db)")
}
