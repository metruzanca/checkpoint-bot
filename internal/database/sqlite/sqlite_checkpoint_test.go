package sqlite

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *SqliteDatabase {
	t.Helper()

	// Use in-memory database
	db := NewSqliteDatabase(":memory:")
	require.NotNil(t, db, "Failed to create test database")

	return db
}

// TestCreateCheckpoint tests creating a checkpoint
func TestTemplate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	// ctx := context.Background()
}
