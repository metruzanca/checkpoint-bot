package sqlite

import (
	"context"

	"github.com/charmbracelet/log"
	"github.com/metruzanca/checkpoint-bot/internal/database/queries"
)

// Implementations for CheckpointDatabase interface

func (db *SqliteDatabase) CreateCheckpoint(ctx context.Context, params queries.CreateCheckpointParams) (*queries.Checkpoint, error) {
	record, err := db.queries.CreateCheckpoint(ctx, params)
	if err != nil {
		return nil, err
	}

	log.Info("Created checkpoint", "id", record.ID, "channel_id", record.ChannelID, "guild_id", record.GuildID, "discord_user", record.DiscordUser, "scheduled_at", record.ScheduledAt)

	return &record, nil
}

func (db *SqliteDatabase) GetUpcomingCheckpoints(ctx context.Context) ([]queries.Checkpoint, error) {
	records, err := db.queries.GetUpcomingCheckpoints(ctx)
	if err != nil {
		return nil, err
	}
	return records, nil
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
	log.Info("Created goal", "goal_id", record.ID, "discord_user", record.DiscordUser, "checkpoint_id", record.CheckpointID)
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

func (db *SqliteDatabase) GetGuild(ctx context.Context, guildID string) (*queries.Guild, error) {
	record, err := db.queries.GetGuild(ctx, guildID)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (db *SqliteDatabase) CreateGuild(ctx context.Context, params queries.CreateGuildParams) (*queries.Guild, error) {
	record, err := db.queries.CreateGuild(ctx, params)
	if err != nil {
		return nil, err
	}
	log.Info("Created guild", "guild_id", record.GuildID, "timezone", record.Timezone, "owner_id", record.OwnerID)
	return &record, nil
}

func (db *SqliteDatabase) GetCheckpointByScheduledAtAndChannel(ctx context.Context, params queries.GetCheckpointByScheduledAtAndChannelParams) (*queries.Checkpoint, error) {
	record, err := db.queries.GetCheckpointByScheduledAtAndChannel(ctx, params)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (db *SqliteDatabase) GetPastCheckpointsByChannel(ctx context.Context, channelID string) ([]queries.Checkpoint, error) {
	records, err := db.queries.GetPastCheckpointsByChannel(ctx, channelID)
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (db *SqliteDatabase) GetUpcomingCheckpointByGuildAndChannel(ctx context.Context, params queries.GetUpcomingCheckpointByGuildAndChannelParams) (*queries.Checkpoint, error) {
	record, err := db.queries.GetUpcomingCheckpointByGuildAndChannel(ctx, params)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (db *SqliteDatabase) GetGoalByCheckpointAndUser(ctx context.Context, params queries.GetGoalByCheckpointAndUserParams) (*queries.Goal, error) {
	record, err := db.queries.GetGoalByCheckpointAndUser(ctx, params)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (db *SqliteDatabase) UpdateGoalDescription(ctx context.Context, params queries.UpdateGoalDescriptionParams) error {
	err := db.queries.UpdateGoalDescription(ctx, params)
	if err != nil {
		return err
	}
	log.Info("Updated goal description", "checkpoint_id", params.CheckpointID, "discord_user", params.DiscordUser)
	return nil
}

func (db *SqliteDatabase) UpdateGoalStatus(ctx context.Context, params queries.UpdateGoalStatusParams) error {
	err := db.queries.UpdateGoalStatus(ctx, params)
	if err != nil {
		return err
	}
	log.Info("Updated goal status", "checkpoint_id", params.CheckpointID, "discord_user", params.DiscordUser, "status", params.Status)
	return nil
}

func (db *SqliteDatabase) GetUpcomingCheckpointsByGuildAndChannel(ctx context.Context, params queries.GetUpcomingCheckpointsByGuildAndChannelParams) ([]queries.Checkpoint, error) {
	records, err := db.queries.GetUpcomingCheckpointsByGuildAndChannel(ctx, params)
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (db *SqliteDatabase) GetGoalsByCheckpoint(ctx context.Context, checkpointID int64) ([]queries.Goal, error) {
	records, err := db.queries.GetGoalsByCheckpoint(ctx, checkpointID)
	if err != nil {
		return nil, err
	}
	return records, nil
}
