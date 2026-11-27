package preacher_test

import (
	"context"
	"fmt"
	"testing"
	"time"
	"uuid"

	"github.com/bryanfks-dev/logos/internal/repository/preacher"
	"github.com/bryanfks-dev/logos/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPreacherTestDB(t *testing.T) *gorm.DB {
	dbName := fmt.Sprintf(
		"file:preacher_repo_%s?mode=memory&cache=shared",
		uuid.NewV7().String(),
	)
	testDB, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, testDB.AutoMigrate(&entity.Preacher{}))

	return testDB
}

func TestPreacherRepository(t *testing.T) {
	ctx := context.Background()
	testDB := setupPreacherTestDB(t)
	testLog := zap.NewNop()

	repo := preacher.NewPreacherRepository(testDB, testLog)

	p := &entity.Preacher{Name: "John Doe", ChurchName: "Grace Church"}

	assert.NoError(t, repo.Create(ctx, p))

	p.ChurchName = "Grace Community Church"

	assert.NoError(t, repo.Update(ctx, p))

	fetched, err := repo.FindByID(ctx, p.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Grace Community Church", fetched.ChurchName)

	p2 := &entity.Preacher{Name: "Jane Smith"}

	assert.NoError(t, repo.Create(ctx, p2))

	prefixFilter := entity.PreacherFilter{NamePrefix: "Jo"}
	filtered, err := repo.FindAll(ctx, prefixFilter)
	assert.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "John Doe", filtered[0].Name)

	assert.NoError(t, repo.Delete(ctx, p.ID))

	emptyFilter := entity.PreacherFilter{}
	filteredAfterDelete, err := repo.FindAll(ctx, emptyFilter)
	assert.NoError(t, err)
	assert.Len(t, filteredAfterDelete, 1)
}

func TestPreacherRepository_FindUpdatedSinceAndBatch(t *testing.T) {
	ctx := context.Background()
	testDB := setupPreacherTestDB(t)
	testLog := zap.NewNop()
	repo := preacher.NewPreacherRepository(testDB, testLog)

	now := time.Now().UTC()
	future := now.Add(2 * time.Hour)

	pActive := &entity.Preacher{
		ID:        "p-act",
		Name:      "Active",
		UpdatedAt: future,
	}
	pDeleted := &entity.Preacher{
		ID:        "p-del",
		Name:      "Deleted",
		UpdatedAt: future,
		DeletedAt: gorm.DeletedAt{Time: future, Valid: true},
	}
	require.NoError(t, testDB.Create(pActive).Error)
	require.NoError(t, testDB.Create(pDeleted).Error)

	results, err := repo.FindUpdatedSince(ctx, now)
	require.NoError(t, err)
	var ids []string
	for _, r := range results {
		ids = append(ids, r.ID)
	}
	assert.Contains(t, ids, "p-act")
	assert.Contains(t, ids, "p-del")

	require.NoError(t, repo.CreateBatch(ctx, nil))
	require.NoError(t, repo.UpdateBatch(ctx, nil))

	items := []*entity.Preacher{
		{ID: "pb-1", Name: "Preacher 1"},
		{ID: "pb-2", Name: "Preacher 2"},
	}
	require.NoError(t, repo.CreateBatch(ctx, items))

	items[0].Name = "Preacher 1 Updated"
	require.NoError(t, repo.UpdateBatch(ctx, items))

	fetched, err := repo.FindByIDs(ctx, []string{"pb-1", "pb-2"})
	require.NoError(t, err)
	require.Len(t, fetched, 2)
}
