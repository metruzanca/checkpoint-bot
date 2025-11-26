package migrations

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var embedMigrations embed.FS
var embededMigrationsDir string = "."
var dialect string = "sqlite3"

var localMigrationsDir string = "internal/migrations"

// errorOnlyLogger logs only errors (Fatalf) and suppresses normal messages (Printf)
type errorOnlyLogger struct{}

func (l *errorOnlyLogger) Printf(format string, v ...interface{}) {
	// Suppress normal messages
}

func (l *errorOnlyLogger) Fatalf(format string, v ...interface{}) {
	log.Fatal("Migration error", "message", fmt.Sprintf(format, v...))
}

func init() {
	goose.SetBaseFS(embedMigrations)
	goose.SetLogger(&errorOnlyLogger{})

	if err := goose.SetDialect(dialect); err != nil {
		log.Fatal("Failed to set goose dialect", "err", err)
	}
}

func Up(db *sql.DB) error {
	return goose.Up(db, embededMigrationsDir)
}

func Down(db *sql.DB) error {
	return goose.Down(db, embededMigrationsDir)
}

func Status(db *sql.DB) error {
	return goose.Status(db, embededMigrationsDir)
}

func Create(db *sql.DB, name string) error {
	return goose.Create(db, localMigrationsDir, name, "sql")
}
