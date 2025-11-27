## Conventions

We're building a 12 Factor app.

Logging is important. Especially in this stage as we're going to need to debug things. Take advantage of charmbracelet/log's methods.

## Dependencies

This project uses github.com/charmbracelet/log instead of go std log.
Logging values need to be prefixed with a string name e.g. `log.Info("My log", "value", value)`.

## Architecture & Business Logic

Business logic lives in two main areas:

- **Data layer**: `internal/database/queries/` and `internal/database/migrations/` for database operations
- **User interactions**: `internal/server/commands/` for Discord command handlers

### Database Layer (`internal/database/`)

#### Queries (`internal/database/queries/`)

- Uses **sqlc** for type-safe SQL queries
- SQL queries are defined in `queries.sql` with sqlc annotations:
  - `-- name: QueryName :one` - Returns a single row
  - `-- name: QueryName :exec` - Executes without returning rows
- **Convention**: All Create operations return the created record (`:one`)
- Update/Delete operations use `:exec` (no return value)
- Models are auto-generated from SQL schema by sqlc
- The `Queries` struct wraps database operations and supports transactions via `WithTx()`
- **DO NOT EDIT** generated files (`queries.sql.go`, `models.go`, `db.go`)

#### Database Interface Pattern

- `CheckpointDatabase` interface in `checkpoint_db.go` defines all database operations
- Implementations (e.g., `SQLiteCheckpointDatabase`) wrap sqlc-generated queries
- Interface methods return pointers (`*queries.Checkpoint`) while sqlc returns values
- Database implementations convert sqlc return values to pointers for interface consistency
- Migrations run automatically on database initialization

#### Migrations (`internal/database/migrations/`)

- Uses **goose** for database migrations
- Migrations are embedded using `//go:embed *.sql`
- Migration files follow pattern: `NNNNN_name.sql` (e.g., `00001_initial.sql`)
- Each migration must have:
  - `-- +goose Up` section for forward migration
  - `-- +goose Down` section for rollback
- Custom logger (`errorOnlyLogger`) only logs errors, suppresses normal messages
- SQLite3 dialect is configured

### Commands (`internal/server/commands/`)

#### Command Registration Pattern

- Commands are registered via `registerCommand()` in `init()` functions
- All commands are stored in a global `commands` map
- Each command is a `Command` struct containing:
  - `ApplicationCommand` (Discord command definition)
  - `Handler` function (business logic)
- Handler signature: `func(db database.CheckpointDatabase, s *discordgo.Session, i *discordgo.InteractionCreate)`
- Commands are registered per guild on bot startup
- Unregistered commands are automatically cleaned up

#### Command Handler Pattern

- Use `context.Background()` for database operations
- Error handling uses `ErrorResponse()` helper function for consistent error messages
- Log errors with context: `log.Error("cannot create checkpoint", "err", err, "channel", i.ChannelID)`
- Respond to interactions using `s.InteractionRespond()`
