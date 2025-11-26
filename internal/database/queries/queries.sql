/*
  File Conventions:
    - All Create operations return the created record
*/

-- name: CreateCheckpoint :one
INSERT INTO checkpoints (scheduled_at, channel_id)
VALUES (?, ?) RETURNING *;

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

-- name: GetAllStats :one
SELECT 
    COUNT(DISTINCT c.id) as total_checkpoints,
    COUNT(DISTINCT g.id) as total_goals,
    COUNT(DISTINCT CASE WHEN g.status = 'completed' THEN g.id END) as completed_goals,
    COUNT(DISTINCT CASE WHEN g.status = 'failed' THEN g.id END) as failed_goals,
    COUNT(DISTINCT CASE WHEN g.status = 'pending' THEN g.id END) as pending_goals,
    COUNT(DISTINCT a.discord_user) as unique_participants,
    COUNT(DISTINCT a.id) as total_attendance
FROM checkpoints c
LEFT JOIN goals g ON c.id = g.checkpoint_id
LEFT JOIN attendance a ON c.id = a.checkpoint_id;

-- name: GetUserStats :one
SELECT 
    COALESCE((SELECT COUNT(*) FROM attendance a WHERE a.discord_user = sqlc.arg(discord_user)), 0) as user_checkpoints_attended,
    COALESCE(COUNT(g.id), 0) as user_total_goals,
    COALESCE(COUNT(CASE WHEN g.status = 'completed' THEN 1 END), 0) as user_completed_goals,
    COALESCE(COUNT(CASE WHEN g.status = 'failed' THEN 1 END), 0) as user_failed_goals,
    COALESCE(COUNT(CASE WHEN g.status = 'pending' THEN 1 END), 0) as user_pending_goals
FROM (SELECT sqlc.arg(discord_user) as discord_user) u
LEFT JOIN goals g ON g.discord_user = u.discord_user;

-- name: GetUpcomingCheckpoint :one
SELECT * FROM checkpoints
WHERE datetime(scheduled_at) >= datetime('now')
ORDER BY datetime(scheduled_at) ASC
LIMIT 1;

-- name: MarkAttendance :exec
INSERT OR IGNORE INTO attendance (discord_user, checkpoint_id)
VALUES (?, ?);

