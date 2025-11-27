package cmd

import (
	"database/sql"

	"github.com/charmbracelet/log"
	"github.com/metruzanca/checkpoint-bot/internal/config"
	"github.com/metruzanca/checkpoint-bot/internal/database/migrations"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "modernc.org/sqlite"
)

var migrateCmd = &cobra.Command{
	Use:               "migrate",
	Short:             "Run database migrations",
	PersistentPreRunE: config.PersistentPreRunE,
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

		// Open database connection
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			log.Fatal("Failed to open database", "err", err, "path", dbPath)
		}
		defer db.Close()

		// Set SQLite PRAGMA statements before running migrations
		// These must be set outside of transactions (goose runs migrations in transactions)
		// Using db.Exec() instead of db.Conn() to avoid deadlock issues
		// Enable WAL (Write-Ahead Logging) mode for better concurrency
		if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
			log.Fatal("Error setting journal_mode", "err", err)
		}
		// Enable foreign key constraints for SQLite
		if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
			log.Fatal("Error enabling foreign keys", "err", err)
		}
		// Set busy timeout to handle concurrent access gracefully
		if _, err := db.Exec("PRAGMA busy_timeout=5000;"); err != nil {
			log.Fatal("Error setting busy_timeout", "err", err)
		}

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
	migrateCmd.Flags().String("DB_PATH", "./db/checkpoint.db", "Path to SQLite database file (default: ./db/checkpoint.db)")
	rootCmd.AddCommand(migrateCmd)
}
