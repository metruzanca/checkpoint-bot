# Project Conventions and Architecture Guide

## Core Principles

- **12 Factor App**: This project follows 12 Factor app principles
- **Logging**: Use structured logging extensively for debugging. Use `github.com/charmbracelet/log` (NOT standard Go log)
- **Business Logic Separation**: Business logic lives in two distinct layers:
  - Data layer: `internal/database/queries/` and `internal/database/migrations/`
  - User interactions: `internal/server/commands/`

---

## Logging Rules

### DO:

- Use `github.com/charmbracelet/log` package
- Prefix all log values with string names: `log.Info("message", "key", value)`
- Include context in error logs: `log.Error("cannot create checkpoint", "err", err, "channel", i.ChannelID)`
- Use appropriate log levels: `log.Debug()`, `log.Info()`, `log.Error()`, `log.Fatal()`

### DON'T:

- Use standard Go `log` package
- Log without context or key-value pairs

### Example:

```go
log.Info("Command executed", "command", cmdName, "user", userID)
log.Error("cannot create checkpoint", "err", err, "channel", i.ChannelID)
```

---

## Database Layer (`internal/database/`)

### SQL Queries (`internal/database/queries/`)

**Tool**: Uses **sqlc** for type-safe SQL queries

**File Location**: `internal/database/queries/queries.sql`

**Query Annotations**:

- `-- name: QueryName :one` - Returns a single row (use for SELECT, INSERT with RETURNING)
- `-- name: QueryName :exec` - Executes without returning rows (use for UPDATE, DELETE, INSERT without RETURNING)

**Conventions**:

- **ALL Create operations MUST return the created record** (`:one` annotation)
- Update/Delete operations use `:exec` (no return value)
- Models are auto-generated from SQL schema by sqlc
- The `Queries` struct wraps database operations and supports transactions via `WithTx()`

**CRITICAL RULES**:

- **DO NOT EDIT** generated files: `queries.sql.go`, `models.go`, `db.go`
- Edit only `queries.sql` - sqlc generates the Go code
- After editing `queries.sql`, run sqlc to regenerate Go code

**Example Query**:

```sql
-- name: CreateCheckpoint :one
INSERT INTO checkpoints (scheduled_at, channel_id)
VALUES (?, ?) RETURNING *;

-- name: CompleteGoal :exec
UPDATE goals
SET status = 'completed'
WHERE discord_user = ? AND checkpoint_id = ?;
```

### Database Interface Pattern

**Interface Location**: `internal/database/checkpoint_db.go`

**Pattern**:

- `CheckpointDatabase` interface defines all database operations
- Implementations (e.g., `SQLiteCheckpointDatabase` in `internal/database/sqlite/sqlite_db.go`) wrap sqlc-generated queries
- **Important**: Interface methods return pointers (`*queries.Checkpoint`) while sqlc returns values
- Database implementations MUST convert sqlc return values to pointers for interface consistency

**Example Implementation Pattern**:

```go
func (d *SqliteDatabase) CreateCheckpoint(ctx context.Context, params queries.CreateCheckpointParams) (*queries.Checkpoint, error) {
    result, err := d.queries.CreateCheckpoint(ctx, params)
    if err != nil {
        return nil, err
    }
    return &result, nil  // Convert value to pointer
}
```

### Migrations (`internal/database/migrations/`)

**Tool**: Uses **goose** for database migrations

**File Pattern**: `NNNNN_name.sql` (e.g., `00001_initial.sql`)

**Migration Structure**:

- Each migration file MUST have:
  - `-- +goose Up` section for forward migration
  - `-- +goose Down` section for rollback
- Migrations are embedded using `//go:embed *.sql`
- Custom logger (`errorOnlyLogger`) only logs errors, suppresses normal messages
- SQLite3 dialect is configured
- Migrations run automatically on database initialization

**Example Migration**:

```sql
-- +goose Up
CREATE TABLE checkpoints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scheduled_at TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE checkpoints;
```

---

## Commands (`internal/server/commands/`)

### Command Registration Pattern

**Registration Location**: Each command file has an `init()` function that calls `registerCommand()`

**Global Storage**: All commands are stored in a global `commands` map (defined in `commands.go`)

**Command Structure**:

```go
type Command struct {
    discordgo.ApplicationCommand  // Discord command definition
    Handler func(db database.CheckpointDatabase, s *discordgo.Session, i *discordgo.InteractionCreate)
}
```

**Registration Pattern**:

```go
var MyCommand = &Command{
    ApplicationCommand: discordgo.ApplicationCommand{
        Name:        "command-name",
        Description: "Command description",
        Options: []*discordgo.ApplicationCommandOption{...},
    },
    Handler: func(db database.CheckpointDatabase, s *discordgo.Session, i *discordgo.InteractionCreate) {
        // Handler implementation
    },
}

func init() {
    registerCommand(MyCommand)
}
```

**Lifecycle**:

- Commands are registered per guild on bot startup
- Unregistered commands are automatically cleaned up via `clearUnregisteredCommands()`
- Commands are registered via `CommandHandler.RegisterCommands()` in bot startup

### Command Handler Pattern

**Handler Signature**:

```go
func(db database.CheckpointDatabase, s *discordgo.Session, i *discordgo.InteractionCreate)
```

**Required Patterns**:

1. **Context Usage**:

   - Use `context.Background()` for all database operations
   - Example: `db.CreateCheckpoint(context.Background(), params)`

2. **Error Handling**:

   - Use `ErrorResponse()` helper function for consistent error messages
   - Always log errors with context before responding
   - Pattern: `log.Error("action description", "err", err, "context_key", contextValue)`

3. **Interaction Response**:
   - Use `s.InteractionRespond()` to respond to interactions
   - Use `ErrorResponse()` helper for error responses
   - Always respond to interactions (don't leave them hanging)

**Example Handler**:

```go
Handler: func(db database.CheckpointDatabase, s *discordgo.Session, i *discordgo.InteractionCreate) {
    result, err := db.SomeOperation(context.Background(), params)
    if err != nil {
        log.Error("cannot perform operation", "err", err, "channel", i.ChannelID)
        ErrorResponse(s, i, "Error message for user")
        return
    }

    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: "Success message",
        },
    })
}
```

**Error Response Helper**:

- Location: `internal/server/commands/commands.go`
- Function: `ErrorResponse(s *discordgo.Session, i *discordgo.InteractionCreate, message string)`
- Use this for all error responses to maintain consistency

---

## File Organization Summary

### Database Files

- `internal/database/checkpoint_db.go` - Database interface definition
- `internal/database/queries/queries.sql` - SQL queries (EDIT THIS)
- `internal/database/queries/queries.sql.go` - Generated by sqlc (DO NOT EDIT)
- `internal/database/queries/models.go` - Generated by sqlc (DO NOT EDIT)
- `internal/database/queries/db.go` - Generated by sqlc (DO NOT EDIT)
- `internal/database/migrations/NNNNN_name.sql` - Migration files
- `internal/database/sqlite/sqlite_db.go` - SQLite implementation

### Command Files

- `internal/server/commands/commands.go` - Command registration infrastructure
- `internal/server/commands/*.go` - Individual command implementations
- Each command file MUST have an `init()` function calling `registerCommand()`

### Server Files

- `internal/server/bot.go` - Bot initialization and lifecycle
- Commands are registered via `CommandHandler.RegisterCommands()` in `Bot.Start()`

---

## Quick Reference Checklist

When creating a new command:

- [ ] Create command file in `internal/server/commands/`
- [ ] Define `Command` struct with `ApplicationCommand` and `Handler`
- [ ] Use `context.Background()` for database calls
- [ ] Use `ErrorResponse()` for error handling
- [ ] Log errors with context: `log.Error("action", "err", err, "key", value)`
- [ ] Call `registerCommand()` in `init()` function
- [ ] Always respond to interactions using `s.InteractionRespond()`

When creating a new database query:

- [ ] Add query to `internal/database/queries/queries.sql`
- [ ] Use `:one` for Create operations (must return created record)
- [ ] Use `:exec` for Update/Delete operations
- [ ] Run sqlc to regenerate Go code
- [ ] Add method to `CheckpointDatabase` interface in `checkpoint_db.go`
- [ ] Implement method in `internal/database/sqlite/sqlite_db.go` (convert value to pointer)

When creating a new migration:

- [ ] Create file `NNNNN_name.sql` in `internal/database/migrations/`
- [ ] Include `-- +goose Up` section
- [ ] Include `-- +goose Down` section
- [ ] Migration runs automatically on database initialization
