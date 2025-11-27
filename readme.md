<div align="center">
  <img src=".github/Remote_Caffeinator.png" alt="Risk of Rain 2's remote caffeinator" width="200">
  
  # ☕ Caffeinator (Checkpoint Bot)
  
  **A Discord bot for accountability and goal tracking**
  
  [![Go Version](https://img.shields.io/badge/go-1.25-blue.svg)](https://golang.org)
  [![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
  
</div>

---

## 📖 About

**Caffeinator** is a Discord bot designed to help accountability groups stay on track. Create scheduled checkpoints, set goals, and track your progress together. Perfect for study groups, fitness buddies, or any team that wants to maintain accountability through regular check-ins.

### ✨ Features

- **📅 Checkpoint Management**: Create scheduled checkpoints with date and time support
- **🎯 Goal Tracking**: Set and manage goals for upcoming checkpoints
- **⏰ Timezone Support**: Automatic timezone handling per server
- **📊 Status Tracking**: Mark goals as completed, incomplete, or failed
- **👥 Multi-Server Support**: Works across multiple Discord servers
- **🛡️ Rate Limiting**: Built-in protection against command spam
- **🔐 Admin Controls**: Administrators can manage other users' goals
- **💾 Persistent Storage**: SQLite database for reliable data storage
- **⚡ High Performance & Lightweight**: Built with Go for optimal resource usage. This is NOT a Node.js app that will consume 300MB+ of RAM. Caffeinator idles at just 10-12MB RAM usage, with typical Go API server memory characteristics—efficient, fast, and perfect for long-running deployments.

---

## 🛠️ Tech Stack

- **Language**: [Go](https://golang.org) 1.25+
- **Discord API**: [discordgo](https://github.com/bwmarrin/discordgo)
- **Database**: SQLite with [sqlc](https://sqlc.dev) for type-safe queries
- **Migrations**: [goose](https://github.com/pressly/goose)
- **Logging**: [charmbracelet/log](https://github.com/charmbracelet/log)
- **CLI**: [Cobra](https://github.com/spf13/cobra)

---

## 🚀 Quick Start

### Prerequisites

- Docker (recommended) or a server with Docker support
- A Discord bot token ([Discord Developer Portal](https://discord.com/developers/applications))

### Using Docker (Recommended)

The easiest way to run Caffeinator is with Docker. An official Docker image will be available soon.

**Coming Soon**: Official Docker image will be published to Docker Hub.

For now, you can build and run from source:

```bash
# Clone the repository
git clone https://github.com/metruzanca/checkpoint-bot.git
cd checkpoint-bot

# Build and run with Docker
docker build -t caffeinator .
docker run -d \
  --name caffeinator \
  -e TOKEN="your-bot-token" \
  -v caffeinator-data:/app/db \
  caffeinator
```

### Configuration

The bot requires only one configuration option:

- `TOKEN` - Discord bot token (required)

Set it as an environment variable when running Docker:

```bash
docker run -e TOKEN="your-bot-token" caffeinator
```

### Invite the Bot to Your Server

1. Get your bot's Client ID from the [Discord Developer Portal](https://discord.com/developers/applications)
2. Use this invite link (replace `YOUR_CLIENT_ID` with your actual Client ID):
   ```
   https://discord.com/api/oauth2/authorize?client_id=YOUR_CLIENT_ID&permissions=0&scope=bot%20applications.commands
   ```
3. Select your server and authorize the bot

---

## 💻 For Developers

### Building from Source

If you want to build from source or contribute:

1. **Clone the repository**

   ```bash
   git clone https://github.com/metruzanca/checkpoint-bot.git
   cd checkpoint-bot
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Generate database code** (after editing queries)

   ```bash
   go generate ./...
   ```

4. **Build the bot**

   ```bash
   go build -o checkpoint ./cmd/...
   ```

5. **Run the bot**
   ```bash
   ./checkpoint --TOKEN="your-bot-token"
   ```

### Advanced Configuration

For development or advanced deployments, additional options are available:

- `CLIENT_ID` - Discord application client ID (for invite links)
- `DB_PATH` - Path to SQLite database file (default: `./db/checkpoint.db`)
- `CHANNEL_ID` - Optional channel ID for startup notifications
- `STARTUP_MESSAGE` - Enable/disable startup messages (default: `true`)

---

## 📚 Usage

### Commands

#### `/checkpoint`

Create a new checkpoint for your accountability group.

**Options:**

- `date` (required): Date in `YYYY-MM-DD` format
- `time` (required): Time in `HH:MM` or `H:MM AM/PM` format

**Example:**

```
/checkpoint date:2024-01-15 time:7:00 PM
```

#### `/goal`

Set or edit your goals for the upcoming checkpoint.

**Options:**

- `user` (optional): User whose goals to edit (admin only)
- `status` (optional): Set goal status (`completed`, `incomplete`, `failed`)

**Example:**

```
/goal
/goal status:completed
```

#### `/next`

View the next upcoming checkpoint and associated goals.

---

## 🏗️ Project Structure

```
checkpoint-bot/
├── cmd/                    # CLI commands (root, invite, clear)
├── internal/
│   ├── config/            # Configuration management
│   ├── database/
│   │   ├── migrations/    # Database migrations (goose)
│   │   ├── queries/       # SQL queries (sqlc)
│   │   ├── checkpoint_db.go  # Database interface
│   │   └── sqlite/        # SQLite implementation
│   ├── server/
│   │   ├── bot.go         # Bot initialization and lifecycle
│   │   └── commands/      # Discord slash command handlers
│   │       ├── checkpoint.go
│   │       ├── goals.go
│   │       └── commands.go
│   └── util/              # Utility functions
└── main.go                # Entry point
```

### Key Architecture Points

- **Database Layer**: SQL queries are defined in `queries.sql` and generated by sqlc
- **Command Handlers**: Each command is self-contained in `internal/server/commands/`
- **Migrations**: Database schema changes go in `internal/database/migrations/`
- **12 Factor App**: Follows 12 Factor app principles for configuration and deployment

---

## 🧪 Development

### Adding a New Command

1. Create a new file in `internal/server/commands/`
2. Define your command with `ApplicationCommand` and `Handler`
3. Register it in the `init()` function using `registerCommand()`
4. Use `dbContext()` helper for database operations (with `defer cancel()`)
5. Handle errors by logging and responding directly with `s.InteractionRespond()`

### Adding a Database Query

1. Add query to `internal/database/queries/queries.sql`
2. Use `:one` annotation for Create operations (must return created record)
3. Use `:exec` annotation for Update/Delete operations
4. Run `go generate ./...` to regenerate sqlc code
5. Add method to `CheckpointDatabase` interface
6. Implement method in SQLite implementation (convert return value to pointer)

### Running Migrations

Migrations run automatically on database initialization. To create a new migration:

1. Create file `NNNNN_name.sql` in `internal/database/migrations/`
2. Include both `-- +goose Up` and `-- +goose Down` sections
3. Migration runs automatically on next bot startup

### Building

```bash
# Verify the project builds
go build ./...

# Generate sqlc code (after editing queries.sql)
go generate ./...
```

---

## 📚 Resources

- [Awesome DiscordGo](https://github.com/bwmarrin/discordgo/wiki/Awesome-DiscordGo) - DiscordGo resources and examples
- [Discordgo Examples](https://github.com/bwmarrin/discordgo/tree/master/examples)
  - [Components](https://github.com/bwmarrin/discordgo/tree/master/examples/components)
  - [Modals](https://github.com/bwmarrin/discordgo/tree/master/examples/modals)
  - [Scheduled Events](https://github.com/bwmarrin/discordgo/tree/master/examples/scheduled_events)

---

## 📝 License

This project is licensed under the MIT License.

---

<div align="center">
  <sub>Built with ☕ and ❤️ for accountability groups everywhere</sub>
</div>
