package gathering_test

import (
	"context"
	"fmt"
	"testing"
	"time"
	"uuid"

	"github.com/bryanfks-dev/logos/infra/db/correlation"
	"github.com/bryanfks-dev/logos/internal/repository/gathering"
	"github.com/bryanfks-dev/logos/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupGatheringTestDB(t *testing.T) *gorm.DB {
	dbName := fmt.Sprintf(
		"file:gathering_repo_%s?mode=memory&cache=shared",
		uuid.NewV7().String(),
	)
	testDB, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, testDB.AutoMigrate(&entity.Gathering{}))

	return testDB
}

func TestGatheringRepository(t *testing.T) {
	ctx := context.Background()
	testDB := setupGatheringTestDB(t)
	testLog := zap.NewNop()

	repo := gathering.NewGatheringRepository(testDB, testLog)

	g := &entity.Gathering{Name: "Sunday Service", SortOrder: 2}

	err := repo.Create(ctx, g)
	assert.NoError(t, err)
	assert.NotEmpty(t, g.ID)

	g.Name = "Sunday Morning Service"

	err = repo.Update(ctx, g)
	assert.NoError(t, err)

	fetched, err := repo.FindByID(ctx, g.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Sunday Morning Service", fetched.Name)

	g2 := &entity.Gathering{Name: "Youth Service", SortOrder: 1}

	assert.NoError(t, repo.Create(ctx, g2))

	all, err := repo.FindAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, all, 2)
	assert.Equal(t, "Youth Service", all[0].Name)

	assert.NoError(t, repo.Delete(ctx, g.ID))

	remaining, err := repo.FindAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, remaining, 1)

	_, err = repo.FindByID(ctx, g.ID)
	assert.Error(t, err)
}

func TestGatheringRepository_FindUpdatedSinceAndBatch(t *testing.T) {
	ctx := context.Background()
	testDB := setupGatheringTestDB(t)
	testLog := zap.NewNop()
	repo := gathering.NewGatheringRepository(testDB, testLog)

	now := time.Now().UTC()
	past := now.Add(-2 * time.Hour)
	future := now.Add(2 * time.Hour)

	gActive := &entity.Gathering{
		ID:        "g-act",
		Name:      "Active",
		UpdatedAt: future,
	}
	gDeleted := &entity.Gathering{
		ID:        "g-del",
		Name:      "Deleted",
		UpdatedAt: future,
		DeletedAt: gorm.DeletedAt{Time: future, Valid: true},
	}
	gPast := &entity.Gathering{ID: "g-past", Name: "Past", UpdatedAt: past}

	require.NoError(t, testDB.Create(gActive).Error)
	require.NoError(t, testDB.Create(gDeleted).Error)
	require.NoError(t, testDB.Create(gPast).Error)

	results, err := repo.FindUpdatedSince(ctx, now)
	require.NoError(t, err)
	var ids []string
	for _, r := range results {
		ids = append(ids, r.ID)
	}
	assert.Contains(t, ids, "g-act")
	assert.Contains(t, ids, "g-del")
	assert.NotContains(t, ids, "g-past")

	require.NoError(t, repo.CreateBatch(ctx, nil))
	require.NoError(t, repo.UpdateBatch(ctx, nil))

	items := []*entity.Gathering{
		{ID: "gb-1", Name: "First", SortOrder: 1},
		{ID: "gb-2", Name: "Second", SortOrder: 2},
	}
	require.NoError(t, repo.CreateBatch(ctx, items))

	items[0].Name = "First Updated"
	items[1].Name = "Second Updated"
	require.NoError(t, repo.UpdateBatch(ctx, items))

	fetched, err := repo.FindByIDs(ctx, []string{"gb-1", "gb-2"})
	require.NoError(t, err)
	require.Len(t, fetched, 2)
}

func TestTransactionPropagation(t *testing.T) {
	ctx := context.Background()
	testDB := setupGatheringTestDB(t)
	testLog := zap.NewNop()

	repo := gathering.NewGatheringRepository(testDB, testLog)

	tx := testDB.Begin()
	txCtx := correlation.WithTx(ctx, tx)

	item := &entity.Gathering{Name: "Tx Rollback Gathering"}
	assert.NoError(t, repo.Create(txCtx, item))

	tx.Rollback()

	all, err := repo.FindAll(ctx)
	assert.NoError(t, err)

	for _, g := range all {
		assert.NotEqual(t, "Tx Rollback Gathering", g.Name)
	}
}
