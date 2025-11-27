# Project Conventions and Architecture Guide

## Quick Summary for AI Agents

**Critical Rules**:

- **MUST**: Use `github.com/charmbracelet/log` for logging (NOT standard Go log)
- **MUST**: Use structured logging with key-value pairs: `log.Info("message", "key", value)`
- **MUST**: Edit only `queries.sql` for database queries (DO NOT edit generated files)
- **MUST**: Use `dbContext()` helper for all database operations in command handlers
- **MUST**: Always respond to Discord interactions using `s.InteractionRespond()`
- **MUST**: Each command file have `init()` function calling `registerCommand()`
- **MUST**: Convert sqlc return values to pointers in database implementations

**Key File Locations**:

- Commands: `internal/server/commands/*.go`
- Database queries: `internal/database/queries/queries.sql` (EDIT THIS)
- Database interface: `internal/database/checkpoint_db.go`
- Database implementation: `internal/database/sqlite/sqlite_db.go`
- Migrations: `internal/database/migrations/NNNNN_name.sql`
- Context helper: `dbContext()` in `internal/server/commands/commands.go`

**Build Commands**:

- **MUST**: Use `go build ./...` to verify the project builds correctly
- **MUST**: Use `go generate ./...` to regenerate sqlc code after editing `queries.sql`

---

## Core Principles

- **12 Factor App**: This project follows 12 Factor app principles
- **Logging**: Use structured logging extensively for debugging. Use `github.com/charmbracelet/log` (NOT standard Go log)
- **Business Logic Separation**: Business logic lives in two distinct layers:
  - Data layer: `internal/database/queries/` and `internal/database/migrations/`
  - User interactions: `internal/server/commands/`

---

## Logging Rules

### REQUIRED: Use Structured Logging

**MUST**: Use `github.com/charmbracelet/log` package for all logging operations.

**MUST NOT**: Use standard Go `log` package.

**MUST**: Prefix all log values with string names using key-value pairs.

**MUST**: Include context in error logs with relevant identifiers.

**MUST**: Use appropriate log levels: `log.Debug()`, `log.Info()`, `log.Error()`, `log.Fatal()`

**MUST NOT**: Log without context or key-value pairs.

### Logging Pattern

**Format**: `log.Level("message", "key1", value1, "key2", value2)`

**Required Context Keys**:

- Error logs: MUST include `"err"` key with error value
- Command logs: SHOULD include `"command"` and `"user"` or `"channel"` keys
- Database logs: SHOULD include relevant identifiers like `"channel"`, `"checkpoint_id"`, etc.

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

**MUST**: Edit only `internal/database/queries/queries.sql` - sqlc generates all Go code from this file.

**MUST NOT**: Edit generated files: `queries.sql.go`, `models.go`, `db.go`

**MUST**: After editing `queries.sql`, run `go generate ./...` to regenerate sqlc code.

**MUST**: Use `go build ./...` to verify the project builds correctly after making changes.

### Query Annotations

**Pattern**: `-- name: QueryName :annotation`

**Annotations**:

- `:one` - Returns a single row (use for SELECT, INSERT with RETURNING)
- `:exec` - Executes without returning rows (use for UPDATE, DELETE, INSERT without RETURNING)

**Rules**:

- **MUST**: ALL Create operations use `:one` annotation and return the created record
- **MUST**: Update/Delete operations use `:exec` annotation (no return value)
- Models are auto-generated from SQL schema by sqlc
- The `Queries` struct wraps database operations and supports transactions via `WithTx()`

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

**Implementation Location**: `internal/database/sqlite/sqlite_db.go`

**Pattern**:

- `CheckpointDatabase` interface defines all database operations
- Implementations (e.g., `SQLiteCheckpointDatabase`) wrap sqlc-generated queries
- **CRITICAL**: Interface methods return pointers (`*queries.Checkpoint`) while sqlc returns values
- **MUST**: Database implementations convert sqlc return values to pointers for interface consistency

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

**Location**: `internal/database/migrations/`

**MUST**: Each migration file include both sections:

- `-- +goose Up` section for forward migration
- `-- +goose Down` section for rollback

**Implementation Details**:

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

**Location**: Each command file in `internal/server/commands/*.go`

**MUST**: Each command file have an `init()` function that calls `registerCommand()`

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
- Commands are registered via `CommandHandler.RegisterCommands()` in `Bot.Start()` (see `internal/server/bot.go`)

### Command Handler Pattern

**Handler Signature**:

```go
func(db database.CheckpointDatabase, s *discordgo.Session, i *discordgo.InteractionCreate)
```

**Required Patterns**:

1. **Context Usage**:

   - **MUST**: Use `dbContext()` helper function for all database operations
   - **MUST**: Always defer `cancel()` after calling `dbContext()`
   - **Pattern**: 
     ```go
     ctx, cancel := dbContext()
     defer cancel()
     ```
   - Example: `db.CreateCheckpoint(ctx, params)`

2. **Error Handling**:

   - **MUST**: Log errors with context before responding
   - **MUST**: Respond to errors directly using `s.InteractionRespond()` with appropriate error messages
   - **Pattern**: 
     ```go
     if err != nil {
         log.Error("action description", "err", err, "context_key", contextValue)
         s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
             Type: discordgo.InteractionResponseChannelMessageWithSource,
             Data: &discordgo.InteractionResponseData{
                 Content: "User-friendly error message",
             },
         })
         return
     }
     ```

3. **Interaction Response**:
   - **MUST**: Use `s.InteractionRespond()` to respond to interactions
   - **MUST**: Always respond to interactions (never leave them hanging)
   - **MUST**: Use user-friendly error messages (use "server" not "guild" in messages)

4. **User-Facing Terminology**:
   - **MUST**: Use "server" instead of "guild" in all user-facing messages (error messages, responses, etc.)
   - **MUST NOT**: Use "guild" in any text that Discord users will see
   - **Note**: "Guild" is Discord's internal API term, but users refer to them as "servers"
   - **OK**: Use "guild" in code, variable names, log messages, and internal documentation
   - **Example**: Error message should say "Error getting server information" not "Error getting guild information"

**Example Handler**:

```go
Handler: func(db database.CheckpointDatabase, s *discordgo.Session, i *discordgo.InteractionCreate) {
    ctx, cancel := dbContext()
    defer cancel()

    result, err := db.SomeOperation(ctx, params)
    if err != nil {
        log.Error("cannot perform operation", "err", err, "channel", i.ChannelID)
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "Error message for user",
            },
        })
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

**Context Helper**:

- **Location**: `internal/server/commands/commands.go`
- **Function**: `dbContext() (context.Context, context.CancelFunc)`
- **MUST**: Use this helper for all database operations to ensure proper timeout handling
- **MUST**: Always defer `cancel()` after calling `dbContext()`

---

## File Organization Summary

### Database Files

**Editable Files**:

- `internal/database/checkpoint_db.go` - Database interface definition (EDIT THIS)
- `internal/database/queries/queries.sql` - SQL queries (EDIT THIS)
- `internal/database/migrations/NNNNN_name.sql` - Migration files (EDIT THIS)
- `internal/database/sqlite/sqlite_db.go` - SQLite implementation (EDIT THIS)

**Generated Files (DO NOT EDIT)**:

- `internal/database/queries/queries.sql.go` - Generated by sqlc
- `internal/database/queries/models.go` - Generated by sqlc
- `internal/database/queries/db.go` - Generated by sqlc

### Command Files

- `internal/server/commands/commands.go` - Command registration infrastructure and `dbContext()` helper
- `internal/server/commands/*.go` - Individual command implementations
- **MUST**: Each command file have an `init()` function calling `registerCommand()`

### Server Files

- `internal/server/bot.go` - Bot initialization and lifecycle
- Commands are registered via `CommandHandler.RegisterCommands()` in `Bot.Start()`

---

## Quick Reference Checklist

### Creating a New Command

**MUST** complete all steps:

1. Create command file in `internal/server/commands/`
2. Define `Command` struct with `ApplicationCommand` and `Handler` fields
3. **MUST**: Use `dbContext()` helper for all database calls (with `defer cancel()`)
4. **MUST**: Log errors with context: `log.Error("action", "err", err, "key", value)`
5. **MUST**: Respond to errors directly using `s.InteractionRespond()` with user-friendly messages
6. **MUST**: Call `registerCommand()` in `init()` function
7. **MUST**: Always respond to interactions using `s.InteractionRespond()`

### Creating a New Database Query

**MUST** complete all steps:

1. Add query to `internal/database/queries/queries.sql`
2. **MUST**: Use `:one` annotation for Create operations (must return created record)
3. **MUST**: Use `:exec` annotation for Update/Delete operations
4. **MUST**: Run `go generate ./...` to regenerate sqlc code
5. Add method to `CheckpointDatabase` interface in `checkpoint_db.go`
6. **MUST**: Implement method in `internal/database/sqlite/sqlite_db.go` (convert sqlc return value to pointer)
7. **MUST**: Run `go build ./...` to verify the project builds correctly

### Creating a New Migration

**MUST** complete all steps:

1. Create file `NNNNN_name.sql` in `internal/database/migrations/` (follow naming pattern)
2. **MUST**: Include `-- +goose Up` section for forward migration
3. **MUST**: Include `-- +goose Down` section for rollback
4. Migration runs automatically on database initialization
