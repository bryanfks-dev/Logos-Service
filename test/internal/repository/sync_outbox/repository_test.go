package syncoutbox_test

import (
	"context"
	"fmt"
	"testing"
	"time"
	"uuid"

	syncoutbox "github.com/bryanfks-dev/logos/internal/repository/sync_outbox"
	"github.com/bryanfks-dev/logos/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSyncOutboxTestDB(t *testing.T) *gorm.DB {
	dbName := fmt.Sprintf(
		"file:sync_outbox_repo_%s?mode=memory&cache=shared",
		uuid.NewV7().String(),
	)
	testDB, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, testDB.AutoMigrate(&entity.SyncOutbox{}))

	return testDB
}

func TestSyncOutboxRepository(t *testing.T) {
	ctx := context.Background()
	testDB := setupSyncOutboxTestDB(t)
	testLog := zap.NewNop()

	repo := syncoutbox.NewSyncOutboxRepository(testDB, testLog)

	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	entry1 := &entity.SyncOutbox{
		EntityType: "sermon_notes",
		EntityID:   "note-1",
		Operation:  "create",
		QueuedAt:   baseTime.Add(1 * time.Second),
	}

	entry2 := &entity.SyncOutbox{
		EntityType: "gatherings",
		EntityID:   "gathering-1",
		Operation:  "create",
		QueuedAt:   baseTime.Add(2 * time.Second),
	}

	assert.NoError(t, repo.Enqueue(ctx, entry1))
	assert.NoError(t, repo.Enqueue(ctx, entry2))

	batch, err := repo.GetPendingBatch(ctx, 10)
	assert.NoError(t, err)
	assert.Len(t, batch, 2)
	assert.Equal(t, "note-1", batch[0].EntityID)

	syncedIDs := []string{entry1.ID}
	assert.NoError(t, repo.MarkSynced(ctx, syncedIDs))

	remaining, err := repo.GetPendingBatch(ctx, 10)
	assert.NoError(t, err)
	assert.Len(t, remaining, 1)

	assert.NoError(t, repo.ClearSynced(ctx, baseTime.Add(3*time.Second)))

	cleared, err := repo.GetPendingBatch(ctx, 10)
	assert.NoError(t, err)
	assert.Empty(t, cleared)
}
