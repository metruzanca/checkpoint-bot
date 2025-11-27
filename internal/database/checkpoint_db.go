package database

import (
	"context"

	"github.com/metruzanca/checkpoint-bot/internal/database/queries"
)

// Defines all operations used by the checkpoint bot
type CheckpointDatabase interface {
	CreateCheckpoint(ctx context.Context, params queries.CreateCheckpointParams) (*queries.Checkpoint, error)
	GetUpcomingCheckpoints(ctx context.Context) ([]queries.Checkpoint, error)
	MarkAttendance(ctx context.Context, params queries.MarkAttendanceParams) error

	CreateGoal(ctx context.Context, params queries.CreateGoalParams) (*queries.Goal, error)
	CompleteGoal(ctx context.Context, params queries.CompleteGoalParams) error
	FailedGoal(ctx context.Context, params queries.FailedGoalParams) error

	GetGuild(ctx context.Context, guildID string) (*queries.Guild, error)
	CreateGuild(ctx context.Context, params queries.CreateGuildParams) (*queries.Guild, error)

	GetCheckpointByScheduledAtAndChannel(ctx context.Context, params queries.GetCheckpointByScheduledAtAndChannelParams) (*queries.Checkpoint, error)
	GetPastCheckpointsByChannel(ctx context.Context, channelID string) ([]queries.Checkpoint, error)
	GetUpcomingCheckpointByGuildAndChannel(ctx context.Context, params queries.GetUpcomingCheckpointByGuildAndChannelParams) (*queries.Checkpoint, error)

	GetGoalByCheckpointAndUser(ctx context.Context, params queries.GetGoalByCheckpointAndUserParams) (*queries.Goal, error)
	UpdateGoalDescription(ctx context.Context, params queries.UpdateGoalDescriptionParams) error

	GetUpcomingCheckpointsByGuildAndChannel(ctx context.Context, params queries.GetUpcomingCheckpointsByGuildAndChannelParams) ([]queries.Checkpoint, error)
	GetGoalsByCheckpoint(ctx context.Context, checkpointID int64) ([]queries.Goal, error)

	Close() error
}
