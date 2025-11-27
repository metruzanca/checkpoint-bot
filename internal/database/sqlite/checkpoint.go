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
