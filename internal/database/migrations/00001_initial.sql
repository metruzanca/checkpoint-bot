-- +goose Up
-- Schema for checkpoint bot database
-- Note: PRAGMA statements (journal_mode, foreign_keys, busy_timeout) are set
-- in the database connection initialization code, not in migrations, because
-- SQLite cannot change journal_mode within a transaction.

-- Guilds table: stores server metadata
CREATE TABLE IF NOT EXISTS guilds (
    guild_id TEXT PRIMARY KEY,
    timezone TEXT NOT NULL, -- Server timezone (e.g., "America/New_York")
    owner_id TEXT NOT NULL, -- Discord user ID of the guild owner/admin
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Checkpoints table: stores checkpoint events with ISO 8601 datetime string and channel
CREATE TABLE IF NOT EXISTS checkpoints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scheduled_at TEXT NOT NULL, -- ISO 8601 datetime string (e.g., "2024-01-15T14:30:00+05:00")
    channel_id TEXT NOT NULL,
    guild_id TEXT NOT NULL,
    discord_user TEXT NOT NULL, -- Creator of the checkpoint
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (guild_id) REFERENCES guilds(guild_id) ON DELETE CASCADE,
    UNIQUE(scheduled_at, channel_id)
);

-- Goals table: stores user goals associated with checkpoints
CREATE TABLE IF NOT EXISTS goals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    discord_user TEXT NOT NULL,
    description TEXT NOT NULL,
    checkpoint_id INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending', -- 'pending', 'completed', 'failed'
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (checkpoint_id) REFERENCES checkpoints(id) ON DELETE CASCADE
);

-- Attendance table: tracks which users attended which checkpoints
CREATE TABLE IF NOT EXISTS attendance (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    discord_user TEXT NOT NULL,
    checkpoint_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (checkpoint_id) REFERENCES checkpoints(id) ON DELETE CASCADE,
    UNIQUE(discord_user, checkpoint_id)
);

-- Checkpoint RSVP table: tracks which users RSVP'd to which checkpoints
CREATE TABLE IF NOT EXISTS checkpoint_rsvp (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    checkpoint_id INTEGER NOT NULL,
    discord_user TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (checkpoint_id) REFERENCES checkpoints(id) ON DELETE CASCADE,
    UNIQUE(checkpoint_id, discord_user)
);

-- View for discord_users: aggregates all unique Discord user IDs from various tables
CREATE VIEW IF NOT EXISTS discord_users AS
SELECT DISTINCT discord_user FROM (
    SELECT discord_user FROM goals
    UNION
    SELECT discord_user FROM attendance
    UNION
    SELECT discord_user FROM checkpoint_rsvp
    UNION
    SELECT discord_user FROM checkpoints
);

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_goals_checkpoint_id ON goals(checkpoint_id);
CREATE INDEX IF NOT EXISTS idx_goals_discord_user ON goals(discord_user);
CREATE INDEX IF NOT EXISTS idx_goals_status ON goals(status);
CREATE INDEX IF NOT EXISTS idx_attendance_checkpoint_id ON attendance(checkpoint_id);
CREATE INDEX IF NOT EXISTS idx_attendance_discord_user ON attendance(discord_user);
CREATE INDEX IF NOT EXISTS idx_checkpoints_scheduled_at ON checkpoints(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_checkpoints_guild_id ON checkpoints(guild_id);
CREATE INDEX IF NOT EXISTS idx_checkpoint_rsvp_checkpoint_id ON checkpoint_rsvp(checkpoint_id);
CREATE INDEX IF NOT EXISTS idx_checkpoint_rsvp_discord_user ON checkpoint_rsvp(discord_user);

-- +goose Down
-- Rollback migration for initial schema

DROP VIEW IF EXISTS discord_users;
DROP TABLE IF EXISTS checkpoint_rsvp;
DROP TABLE IF EXISTS attendance;
DROP TABLE IF EXISTS goals;
DROP TABLE IF EXISTS checkpoints;
DROP TABLE IF EXISTS guilds;

