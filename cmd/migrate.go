package cmd

import (
	"database/sql"

	"github.com/charmbracelet/log"
	"github.com/metruzanca/checkpoint-bot/internal/database/migrations"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "modernc.org/sqlite"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long: `Run database migrations using goose.
	
Examples:
  checkpoint migrate up        - Run all pending migrations
  checkpoint migrate down      - Rollback the last migration
  checkpoint migrate status    - Show migration status
  checkpoint migrate create NAME - Create a new migration`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatal("Migration command required (up, down, status, create)")
		}

		dbPath := viper.GetString("db-path")
		if dbPath == "" {
			dbPath = "checkpoint.db"
		}

		// Open database connection
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			log.Fatal("Failed to open database", "err", err, "path", dbPath)
		}
		defer db.Close()

		// Set goose dialect
		if err := goose.SetDialect("sqlite3"); err != nil {
			log.Fatal("Failed to set goose dialect", "err", err)
		}

		command := args[0]

		switch command {
		case "up":
			if err := migrations.Up(db); err != nil {
				log.Fatal("Failed to run migrations", "err", err)
			}

		case "down":
			if err := migrations.Down(db); err != nil {
				log.Fatal("Failed to rollback migration", "err", err)
			}
			log.Info("Migration rolled back successfully")

		case "status":
			if err := migrations.Status(db); err != nil {
				log.Fatal("Failed to get migration status", "err", err)
			}

		case "create":
			if len(args) < 2 {
				log.Fatal("Migration name required for create command")
			}
			migrationName := args[1]
			if err := migrations.Create(db, migrationName); err != nil {
				log.Fatal("Failed to create migration", "err", err)
			}
			log.Info("Migration created successfully", "name", migrationName)

		default:
			log.Fatal("Unknown migration command", "command", command, "available", []string{"up", "down", "status", "create"})
		}
	},
}

func init() {
	migrateCmd.Flags().String("db-path", "", "Path to SQLite database file (default: checkpoint.db)")
	viper.BindPFlag("db-path", migrateCmd.Flags().Lookup("db-path"))
	rootCmd.AddCommand(migrateCmd)
}
