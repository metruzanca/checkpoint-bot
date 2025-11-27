package database

import (
	"context"

	"github.com/metruzanca/checkpoint-bot/internal/database/queries"
)

// Defines all operations used by the checkpoint bot
type CheckpointDatabase interface {
	CreateCheckpoint(ctx context.Context, params queries.CreateCheckpointParams) (*queries.Checkpoint, error)
	GetUpcomingCheckpoint(ctx context.Context) (*queries.Checkpoint, error)
	MarkAttendance(ctx context.Context, params queries.MarkAttendanceParams) error

	CreateGoal(ctx context.Context, params queries.CreateGoalParams) (*queries.Goal, error)
	CompleteGoal(ctx context.Context, params queries.CompleteGoalParams) error
	FailedGoal(ctx context.Context, params queries.FailedGoalParams) error

	GetAllStats(ctx context.Context) (*queries.GetAllStatsRow, error)
	GetUserStats(ctx context.Context, discordUser string) (*queries.GetUserStatsRow, error)

	Close() error
}
