package sermonseries_test

import (
	"context"
	"fmt"
	"testing"
	"time"
	"uuid"

	sermonseries "github.com/bryanfks-dev/logos/internal/repository/sermon_series"
	"github.com/bryanfks-dev/logos/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSermonSeriesTestDB(t *testing.T) *gorm.DB {
	dbName := fmt.Sprintf(
		"file:sermon_series_repo_%s?mode=memory&cache=shared",
		uuid.NewV7().String(),
	)
	testDB, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, testDB.AutoMigrate(&entity.SermonSeries{}))

	return testDB
}

func TestSermonSeriesRepository(t *testing.T) {
	ctx := context.Background()
	testDB := setupSermonSeriesTestDB(t)
	testLog := zap.NewNop()

	repo := sermonseries.NewSermonSeriesRepository(testDB, testLog)

	s := &entity.SermonSeries{Name: "Romans"}

	err := repo.Create(ctx, s)
	assert.NoError(t, err)

	s.Description = "Study on Romans"

	assert.NoError(t, repo.Update(ctx, s))

	fetched, err := repo.FindByID(ctx, s.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Study on Romans", fetched.Description)

	all, err := repo.FindAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, all, 1)

	assert.NoError(t, repo.Delete(ctx, s.ID))

	remaining, err := repo.FindAll(ctx)
	assert.NoError(t, err)
	assert.Empty(t, remaining)
}

func TestSermonSeriesRepository_FindUpdatedSinceAndBatch(t *testing.T) {
	ctx := context.Background()
	testDB := setupSermonSeriesTestDB(t)
	testLog := zap.NewNop()
	repo := sermonseries.NewSermonSeriesRepository(testDB, testLog)

	now := time.Now().UTC()
	future := now.Add(2 * time.Hour)

	sActive := &entity.SermonSeries{
		ID:        "s-act",
		Name:      "Active",
		UpdatedAt: future,
	}
	sDeleted := &entity.SermonSeries{
		ID:        "s-del",
		Name:      "Deleted",
		UpdatedAt: future,
		DeletedAt: gorm.DeletedAt{Time: future, Valid: true},
	}
	require.NoError(t, testDB.Create(sActive).Error)
	require.NoError(t, testDB.Create(sDeleted).Error)

	results, err := repo.FindUpdatedSince(ctx, now)
	require.NoError(t, err)
	var ids []string
	for _, r := range results {
		ids = append(ids, r.ID)
	}
	assert.Contains(t, ids, "s-act")
	assert.Contains(t, ids, "s-del")

	require.NoError(t, repo.CreateBatch(ctx, nil))
	require.NoError(t, repo.UpdateBatch(ctx, nil))

	items := []*entity.SermonSeries{
		{ID: "sb-1", Name: "Series 1"},
		{ID: "sb-2", Name: "Series 2"},
	}
	require.NoError(t, repo.CreateBatch(ctx, items))

	items[0].Name = "Series 1 Updated"
	require.NoError(t, repo.UpdateBatch(ctx, items))

	fetched, err := repo.FindByIDs(ctx, []string{"sb-1", "sb-2"})
	require.NoError(t, err)
	require.Len(t, fetched, 2)
}
