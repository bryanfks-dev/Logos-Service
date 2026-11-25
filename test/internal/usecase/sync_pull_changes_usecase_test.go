package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
	"uuid"

	"github.com/bryanfks-dev/logos/infra/fs"
	"github.com/bryanfks-dev/logos/internal/repository"
	"github.com/bryanfks-dev/logos/internal/usecase"
	"github.com/bryanfks-dev/logos/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSyncPullTestEnv(
	t *testing.T,
) (*gorm.DB, *usecase.SyncPullChangesUsecase) {
	dbName := fmt.Sprintf(
		"file:sync_pull_test_%s?mode=memory&cache=shared",
		uuid.NewV7().String(),
	)
	testDB, err := gorm.Open(
		sqlite.Open(dbName),
		&gorm.Config{},
	)
	require.NoError(t, err)

	err = testDB.AutoMigrate(
		&entity.Gathering{},
		&entity.SermonSeries{},
		&entity.Preacher{},
		&entity.Tag{},
		&entity.SermonNote{},
		&entity.NoteVerse{},
		&entity.NoteTag{},
		&entity.NotesFTS{},
		&entity.SyncOutboxEntry{},
	)
	require.NoError(t, err)

	log := zap.NewNop()
	sermonNoteRepository := repository.NewSermonNoteRepository(
		testDB,
		fs.NewManager(),
		log,
	)
	gatheringRepository := repository.NewGatheringRepository(testDB, log)
	sermonSeriesRepository := repository.NewSermonSeriesRepository(testDB, log)
	preacherRepository := repository.NewPreacherRepository(testDB, log)
	tagRepository := repository.NewTagRepository(testDB, log)

	syncPullChangesUsecase := usecase.NewSyncPullChangesUsecase(
		&usecase.SyncPullChangesUsecaseOption{
			SermonNoteRepository:   sermonNoteRepository,
			GatheringRepository:    gatheringRepository,
			SermonSeriesRepository: sermonSeriesRepository,
			PreacherRepository:     preacherRepository,
			TagRepository:          tagRepository,
			Log:                    log,
		},
	)

	return testDB, syncPullChangesUsecase
}

func TestSyncPullChangesUsecase_Call(t *testing.T) {
	testDB, u := setupSyncPullTestEnv(t)
	ctx := context.Background()

	now := time.Now().UTC()
	past := now.Add(-2 * time.Hour)
	future := now.Add(2 * time.Hour)

	// Create records with different updated_at
	gActive := &entity.Gathering{
		ID:        "g-pull-act",
		Name:      "Active",
		UpdatedAt: future,
	}
	gDeleted := &entity.Gathering{
		ID:        "g-pull-del",
		Name:      "Deleted",
		UpdatedAt: future,
		DeletedAt: gorm.DeletedAt{Time: future, Valid: true},
	}
	gOld := &entity.Gathering{ID: "g-pull-old", Name: "Old", UpdatedAt: past}
	require.NoError(t, testDB.Create(gActive).Error)
	require.NoError(t, testDB.Create(gDeleted).Error)
	require.NoError(t, testDB.Create(gOld).Error)

	sActive := &entity.SermonSeries{
		ID:        "s-pull-act",
		Name:      "Series",
		UpdatedAt: future,
	}
	require.NoError(t, testDB.Create(sActive).Error)

	pActive := &entity.Preacher{
		ID:        "p-pull-act",
		Name:      "Preacher",
		UpdatedAt: future,
	}
	require.NoError(t, testDB.Create(pActive).Error)

	tagActive := &entity.Tag{ID: "t-pull-act", Name: "Tag", UpdatedAt: future}
	require.NoError(t, testDB.Create(tagActive).Error)

	noteActive := &entity.SermonNote{
		ID:        "n-pull-act",
		Title:     "Note",
		UpdatedAt: future,
	}
	noteDeleted := &entity.SermonNote{
		ID:        "n-pull-del",
		Title:     "Deleted Note",
		UpdatedAt: future,
		DeletedAt: gorm.DeletedAt{Time: future, Valid: true},
	}
	require.NoError(t, testDB.Create(noteActive).Error)
	require.NoError(t, testDB.Create(noteDeleted).Error)

	t.Run(
		"queries all repositories and returns records modified after since including soft deleted",
		func(t *testing.T) {
			res, err := u.Call(ctx, now)
			require.NoError(t, err)
			require.NotNil(t, res)

			assert.False(t, res.ServerTime.IsZero())

			var gIDs []string
			for _, g := range res.Gatherings {
				gIDs = append(gIDs, g.ID)
			}
			assert.Contains(t, gIDs, "g-pull-act")
			assert.Contains(t, gIDs, "g-pull-del")
			assert.NotContains(t, gIDs, "g-pull-old")

			require.Len(t, res.SermonSeries, 1)
			assert.Equal(t, "s-pull-act", res.SermonSeries[0].ID)

			require.Len(t, res.Preachers, 1)
			assert.Equal(t, "p-pull-act", res.Preachers[0].ID)

			require.Len(t, res.Tags, 1)
			assert.Equal(t, "t-pull-act", res.Tags[0].ID)

			var noteIDs []string
			for _, n := range res.SermonNotes {
				noteIDs = append(noteIDs, n.ID)
			}
			assert.Contains(t, noteIDs, "n-pull-act")
			assert.Contains(t, noteIDs, "n-pull-del")
		},
	)

	t.Run("zero since time returns all records", func(t *testing.T) {
		res, err := u.Call(ctx, time.Time{})
		require.NoError(t, err)
		require.NotNil(t, res)

		var gIDs []string
		for _, g := range res.Gatherings {
			gIDs = append(gIDs, g.ID)
		}
		assert.Contains(t, gIDs, "g-pull-old")
	})
}

func TestSyncPullChangesUsecase_ErrorPropagation(t *testing.T) {
	ctx := context.Background()

	t.Run("Call returns repository errors", func(t *testing.T) {
		testDB, _ := setupSyncPullTestEnv(t)
		log := zap.NewNop()
		expectedErr := errors.New("db query failed")

		opt := &usecase.SyncPullChangesUsecaseOption{
			SermonNoteRepository: &failingSermonNoteRepo{
				SermonNoteRepository: repository.NewSermonNoteRepository(
					testDB,
					fs.NewManager(),
					log,
				),
				findUpdatedSinceErr: expectedErr,
			},
			GatheringRepository: repository.NewGatheringRepository(testDB, log),
			SermonSeriesRepository: repository.NewSermonSeriesRepository(
				testDB,
				log,
			),
			PreacherRepository: repository.NewPreacherRepository(testDB, log),
			TagRepository:      repository.NewTagRepository(testDB, log),
			Log:                log,
		}
		u := usecase.NewSyncPullChangesUsecase(opt)

		res, err := u.Call(ctx, time.Time{})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("Call returns gathering repository errors", func(t *testing.T) {
		testDB, _ := setupSyncPullTestEnv(t)
		log := zap.NewNop()
		expectedErr := errors.New("gathering query failed")

		opt := &usecase.SyncPullChangesUsecaseOption{
			SermonNoteRepository: repository.NewSermonNoteRepository(
				testDB,
				fs.NewManager(),
				log,
			),
			GatheringRepository: &failingGatheringRepo{
				GatheringRepository: repository.NewGatheringRepository(
					testDB,
					log,
				),
				findUpdatedSinceErr: expectedErr,
			},
			SermonSeriesRepository: repository.NewSermonSeriesRepository(
				testDB,
				log,
			),
			PreacherRepository: repository.NewPreacherRepository(testDB, log),
			TagRepository:      repository.NewTagRepository(testDB, log),
			Log:                log,
		}
		u := usecase.NewSyncPullChangesUsecase(opt)

		res, err := u.Call(ctx, time.Time{})
		assert.Nil(t, res)
		assert.ErrorIs(t, err, expectedErr)
	})
}
