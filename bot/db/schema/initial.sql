-- Schema for checkpoint bot database

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

