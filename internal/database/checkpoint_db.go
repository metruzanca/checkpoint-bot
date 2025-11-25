package database

import (
	"context"

	"github.com/metruzanca/checkpoint-bot/bot/db"
)

// Defines all operations used by the checkpoint bot
type CheckpointDatabase interface {
	CreateCheckpoint(ctx context.Context, params db.CreateCheckpointParams) (*db.Checkpoint, error)
	GetUpcomingCheckpoint(ctx context.Context) (*db.Checkpoint, error)
	// MarkAttendance(ctx context.Context, params db.MarkAttendanceParams) error

	// CreateGoal(ctx context.Context, params db.CreateGoalParams) (*db.Goal, error)
	// CompleteGoal(ctx context.Context, params db.CompleteGoalParams) error
	// FailedGoal(ctx context.Context, params db.FailedGoalParams) error

	// GetAllStats(ctx context.Context) (*db.GetAllStatsRow, error)
	// GetUserStats(ctx context.Context, discordUser string) (*db.GetUserStatsRow, error)
}
