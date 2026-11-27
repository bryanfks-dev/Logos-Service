package tag_test

import (
	"context"
	"fmt"
	"testing"
	"time"
	"uuid"

	tagrepo "github.com/bryanfks-dev/logos/internal/repository/tag"
	"github.com/bryanfks-dev/logos/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTagTestDB(t *testing.T) *gorm.DB {
	dbName := fmt.Sprintf(
		"file:tag_repo_%s?mode=memory&cache=shared",
		uuid.NewV7().String(),
	)
	testDB, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(
		t,
		testDB.AutoMigrate(&entity.Tag{}, &entity.NoteTag{}),
	)

	return testDB
}

func TestTagRepository(t *testing.T) {
	ctx := context.Background()
	testDB := setupTagTestDB(t)
	testLog := zap.NewNop()

	repo := tagrepo.NewTagRepository(testDB, testLog)

	tag := &entity.Tag{Name: "Grace"}

	assert.NoError(t, repo.Create(ctx, tag))

	tag.Name = "Grace Updated"

	assert.NoError(t, repo.Update(ctx, tag))

	all, err := repo.FindAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, all, 1)
	assert.Equal(t, 0, all[0].UsageCount)

	assert.NoError(t, repo.Delete(ctx, tag.ID))

	remaining, err := repo.FindAll(ctx)
	assert.NoError(t, err)
	assert.Empty(t, remaining)
}

func TestTagRepository_FindUpdatedSinceAndBatch(t *testing.T) {
	ctx := context.Background()
	testDB := setupTagTestDB(t)
	testLog := zap.NewNop()
	repo := tagrepo.NewTagRepository(testDB, testLog)

	now := time.Now().UTC()
	future := now.Add(2 * time.Hour)

	tagRecent := &entity.Tag{
		ID:        "t-act",
		Name:      "Recent Tag",
		UpdatedAt: future,
	}
	require.NoError(t, testDB.Create(tagRecent).Error)

	results, err := repo.FindUpdatedSince(ctx, now)
	require.NoError(t, err)
	var ids []string
	for _, r := range results {
		ids = append(ids, r.ID)
	}
	assert.Contains(t, ids, "t-act")

	require.NoError(t, repo.CreateBatch(ctx, nil))
	require.NoError(t, repo.UpdateBatch(ctx, nil))

	items := []*entity.Tag{
		{ID: "tb-1", Name: "Batch Tag 1"},
		{ID: "tb-2", Name: "Batch Tag 2"},
	}
	require.NoError(t, repo.CreateBatch(ctx, items))

	items[0].Name = "Batch Tag 1 Updated"
	require.NoError(t, repo.UpdateBatch(ctx, items))

	fetched, err := repo.FindByIDsOrNames(ctx, []string{"tb-1", "tb-2"}, nil)
	require.NoError(t, err)
	require.Len(t, fetched, 2)
}
