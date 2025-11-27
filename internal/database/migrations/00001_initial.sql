-- +goose Up
-- Schema for checkpoint bot database
-- Note: PRAGMA statements (journal_mode, foreign_keys, busy_timeout) are set
-- in the database connection initialization code, not in migrations, because
-- SQLite cannot change journal_mode within a transaction.

-- Checkpoints table: stores checkpoint events with ISO 8601 datetime string and channel
CREATE TABLE IF NOT EXISTS checkpoints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scheduled_at TEXT NOT NULL, -- ISO 8601 datetime string (e.g., "2024-01-15T14:30:00+05:00")
    channel_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
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

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_goals_checkpoint_id ON goals(checkpoint_id);
CREATE INDEX IF NOT EXISTS idx_goals_discord_user ON goals(discord_user);
CREATE INDEX IF NOT EXISTS idx_goals_status ON goals(status);
CREATE INDEX IF NOT EXISTS idx_attendance_checkpoint_id ON attendance(checkpoint_id);
CREATE INDEX IF NOT EXISTS idx_attendance_discord_user ON attendance(discord_user);
CREATE INDEX IF NOT EXISTS idx_checkpoints_scheduled_at ON checkpoints(scheduled_at);

-- +goose Down
-- Rollback migration for initial schema

DROP TABLE IF EXISTS attendance;
DROP TABLE IF EXISTS goals;
DROP TABLE IF EXISTS checkpoints;

