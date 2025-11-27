/*
  File Conventions:
    - All Create operations return the created record
*/

-- name: CreateCheckpoint :one
INSERT INTO checkpoints (scheduled_at, channel_id, guild_id, discord_user)
VALUES (?, ?, ?, ?) RETURNING *;

-- name: CreateGoal :one
INSERT INTO goals (discord_user, description, checkpoint_id)
VALUES (?, ?, ?) RETURNING *;

-- name: CompleteGoal :exec
UPDATE goals
SET status = 'completed'
WHERE discord_user = ? AND checkpoint_id = ?;

-- name: FailedGoal :exec
UPDATE goals
SET status = 'failed'
WHERE discord_user = ? AND checkpoint_id = ?;

-- name: GetUpcomingCheckpoints :many
SELECT * FROM checkpoints
WHERE datetime(scheduled_at) >= datetime('now')
ORDER BY datetime(scheduled_at) ASC;

-- name: MarkAttendance :exec
INSERT OR IGNORE INTO attendance (discord_user, checkpoint_id)
VALUES (?, ?);

-- name: GetGuild :one
SELECT * FROM guilds
WHERE guild_id = ?;

-- name: CreateGuild :one
INSERT INTO guilds (guild_id, timezone, owner_id)
VALUES (?, ?, ?) RETURNING *;

-- name: GetCheckpointByScheduledAtAndChannel :one
SELECT * FROM checkpoints
WHERE scheduled_at = ? AND channel_id = ?;

-- name: GetPastCheckpointsByChannel :many
SELECT * FROM checkpoints
WHERE channel_id = ? AND datetime(scheduled_at) < datetime('now')
ORDER BY datetime(scheduled_at) DESC;

-- name: GetUpcomingCheckpointByGuildAndChannel :one
SELECT * FROM checkpoints
WHERE guild_id = ? AND channel_id = ? AND datetime(scheduled_at) >= datetime('now')
ORDER BY datetime(scheduled_at) ASC
LIMIT 1;

