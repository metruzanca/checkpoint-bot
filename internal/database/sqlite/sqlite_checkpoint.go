package sqlite

import (
	"context"

	"github.com/charmbracelet/log"
	"github.com/metruzanca/checkpoint-bot/internal/database/queries"
)

// Implementations for CheckpointDatabase interface

func (db *SqliteDatabase) Close() error {
	return db.db.Close()
}

func (db *SqliteDatabase) CreateCheckpoint(ctx context.Context, params queries.CreateCheckpointParams) (*queries.Checkpoint, error) {
	record, err := db.queries.CreateCheckpoint(ctx, params)
	if err != nil {
		return nil, err
	}

	log.Info("Created checkpoint", "checkpoint", record)

	return &record, nil
}

func (db *SqliteDatabase) GetUpcomingCheckpoint(ctx context.Context) (*queries.Checkpoint, error) {
	record, err := db.queries.GetUpcomingCheckpoint(ctx)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (db *SqliteDatabase) MarkAttendance(ctx context.Context, params queries.MarkAttendanceParams) error {
	err := db.queries.MarkAttendance(ctx, params)
	if err != nil {
		return err
	}
	log.Info("Marked attendance", "discord_user", params.DiscordUser, "checkpoint_id", params.CheckpointID)
	return nil
}

func (db *SqliteDatabase) CreateGoal(ctx context.Context, params queries.CreateGoalParams) (*queries.Goal, error) {
	record, err := db.queries.CreateGoal(ctx, params)
	if err != nil {
		return nil, err
	}
	log.Info("Created goal", "goal", record)
	return &record, nil
}

func (db *SqliteDatabase) CompleteGoal(ctx context.Context, params queries.CompleteGoalParams) error {
	err := db.queries.CompleteGoal(ctx, params)
	if err != nil {
		return err
	}
	log.Info("Completed goal", "discord_user", params.DiscordUser, "checkpoint_id", params.CheckpointID)
	return nil
}

func (db *SqliteDatabase) FailedGoal(ctx context.Context, params queries.FailedGoalParams) error {
	err := db.queries.FailedGoal(ctx, params)
	if err != nil {
		return err
	}
	log.Info("Failed goal", "discord_user", params.DiscordUser, "checkpoint_id", params.CheckpointID)
	return nil
}

func (db *SqliteDatabase) GetAllStats(ctx context.Context) (*queries.GetAllStatsRow, error) {
	record, err := db.queries.GetAllStats(ctx)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (db *SqliteDatabase) GetUserStats(ctx context.Context, discordUser string) (*queries.GetUserStatsRow, error) {
	record, err := db.queries.GetUserStats(ctx, discordUser)
	if err != nil {
		return nil, err
	}
	return &record, nil
}
