package usecase_test

import (
	"context"
	"fmt"
	"testing"
	"time"
	"uuid"

	"github.com/bryanfks-dev/selah/db"
	"github.com/bryanfks-dev/selah/internal/model"
	"github.com/bryanfks-dev/selah/internal/repository"
	"github.com/bryanfks-dev/selah/internal/usecase"
	"github.com/bryanfks-dev/selah/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSyncTestEnv(
	t *testing.T,
) (*gorm.DB, *usecase.SyncUsecase) {
	dbName := fmt.Sprintf(
		"file:sync_test_%s?mode=memory&cache=shared",
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
	txManager := db.NewTxManager(testDB)
	sermonNoteRepository := repository.NewSermonNoteRepository(testDB, log)
	gatheringRepository := repository.NewGatheringRepository(testDB, log)
	sermonSeriesRepository := repository.NewSermonSeriesRepository(testDB, log)
	preacherRepository := repository.NewPreacherRepository(testDB, log)
	tagRepository := repository.NewTagRepository(testDB, log)
	syncOutboxRepository := repository.NewSyncOutboxRepository(testDB, log)

	syncUsecase := usecase.NewSyncUsecase(
		txManager,
		sermonNoteRepository,
		gatheringRepository,
		sermonSeriesRepository,
		preacherRepository,
		tagRepository,
		syncOutboxRepository,
		log,
	)

	return testDB, syncUsecase
}

func TestSyncUsecase_Push_NilBatch(t *testing.T) {
	_, u := setupSyncTestEnv(t)
	ctx := context.Background()

	result, err := u.Push(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, result.SyncedAt, int64(0))
	assert.Empty(t, result.Conflicts)
	assert.Empty(t, result.MissingContentIDs)
}

func TestSyncUsecase_Push_UpsertMetadata(t *testing.T) {
	testDB, u := setupSyncTestEnv(t)
	ctx := context.Background()

	baseTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	olderTime := baseTime.Add(-1 * time.Hour)
	newerTime := baseTime.Add(1 * time.Hour)

	// Pre-populate records in DB with baseTime
	existingGathering := entity.Gathering{
		ID:        "g-1",
		Name:      "Original Gathering",
		ColorHex:  "#000000",
		SortOrder: 1,
		UpdatedAt: baseTime,
	}
	require.NoError(t, testDB.Create(&existingGathering).Error)

	existingSeries := entity.SermonSeries{
		ID:          "s-1",
		Name:        "Original Series",
		Description: "Original description",
		UpdatedAt:   baseTime,
	}
	require.NoError(t, testDB.Create(&existingSeries).Error)

	existingPreacher := entity.Preacher{
		ID:         "p-1",
		Name:       "Original Preacher",
		ChurchName: "Church A",
		UpdatedAt:  baseTime,
	}
	require.NoError(t, testDB.Create(&existingPreacher).Error)

	existingTag := entity.Tag{
		ID:        "t-1",
		Name:      "original-tag",
		UpdatedAt: baseTime,
	}
	require.NoError(t, testDB.Create(&existingTag).Error)

	t.Run(
		"creates new metadata and updates existing when newer",
		func(t *testing.T) {
			batch := &model.SyncBatch{
				Gatherings: []*entity.Gathering{
					{
						ID:        "g-new",
						Name:      "New Gathering",
						ColorHex:  "#111111",
						UpdatedAt: baseTime,
					},
					{
						ID:        "g-1",
						Name:      "Updated Gathering",
						ColorHex:  "#222222",
						UpdatedAt: newerTime,
					},
				},
				Series: []*entity.SermonSeries{
					{
						ID:        "s-new",
						Name:      "New Series",
						UpdatedAt: baseTime,
					},
					{
						ID:        "s-1",
						Name:      "Updated Series",
						UpdatedAt: newerTime,
					},
				},
				Preachers: []*entity.Preacher{
					{
						ID:        "p-new",
						Name:      "New Preacher",
						UpdatedAt: baseTime,
					},
					{
						ID:        "p-1",
						Name:      "Updated Preacher",
						UpdatedAt: newerTime,
					},
				},
				Tags: []*entity.Tag{
					{
						ID:        "t-new",
						Name:      "new-tag",
						UpdatedAt: baseTime,
					},
					{
						ID:        "t-1",
						Name:      "updated-tag",
						UpdatedAt: newerTime,
					},
				},
			}

			result, err := u.Push(ctx, batch)
			require.NoError(t, err)
			assert.NotNil(t, result)

			// Verify creations
			var gNew entity.Gathering
			require.NoError(t, testDB.First(&gNew, "id = ?", "g-new").Error)
			assert.Equal(t, "New Gathering", gNew.Name)

			var sNew entity.SermonSeries
			require.NoError(t, testDB.First(&sNew, "id = ?", "s-new").Error)
			assert.Equal(t, "New Series", sNew.Name)

			var pNew entity.Preacher
			require.NoError(t, testDB.First(&pNew, "id = ?", "p-new").Error)
			assert.Equal(t, "New Preacher", pNew.Name)

			var tNew entity.Tag
			require.NoError(t, testDB.First(&tNew, "id = ?", "t-new").Error)
			assert.Equal(t, "new-tag", tNew.Name)

			// Verify updates
			var gUpdated entity.Gathering
			require.NoError(t, testDB.First(&gUpdated, "id = ?", "g-1").Error)
			assert.Equal(t, "Updated Gathering", gUpdated.Name)

			var sUpdated entity.SermonSeries
			require.NoError(t, testDB.First(&sUpdated, "id = ?", "s-1").Error)
			assert.Equal(t, "Updated Series", sUpdated.Name)

			var pUpdated entity.Preacher
			require.NoError(t, testDB.First(&pUpdated, "id = ?", "p-1").Error)
			assert.Equal(t, "Updated Preacher", pUpdated.Name)

			var tUpdated entity.Tag
			require.NoError(t, testDB.First(&tUpdated, "id = ?", "t-1").Error)
			assert.Equal(t, "updated-tag", tUpdated.Name)
		},
	)

	t.Run(
		"skips updating metadata when incoming UpdatedAt is older",
		func(t *testing.T) {
			batch := &model.SyncBatch{
				Gatherings: []*entity.Gathering{
					{
						ID:        "g-1",
						Name:      "Should Not Overwrite Gathering",
						UpdatedAt: olderTime,
					},
				},
				Series: []*entity.SermonSeries{
					{
						ID:        "s-1",
						Name:      "Should Not Overwrite Series",
						UpdatedAt: olderTime,
					},
				},
				Preachers: []*entity.Preacher{
					{
						ID:        "p-1",
						Name:      "Should Not Overwrite Preacher",
						UpdatedAt: olderTime,
					},
				},
				Tags: []*entity.Tag{
					{
						ID:        "t-1",
						Name:      "should-not-overwrite-tag",
						UpdatedAt: olderTime,
					},
				},
			}

			result, err := u.Push(ctx, batch)
			require.NoError(t, err)
			assert.NotNil(t, result)

			var g entity.Gathering
			require.NoError(t, testDB.First(&g, "id = ?", "g-1").Error)
			assert.Equal(t, "Updated Gathering", g.Name)

			var s entity.SermonSeries
			require.NoError(t, testDB.First(&s, "id = ?", "s-1").Error)
			assert.Equal(t, "Updated Series", s.Name)

			var p entity.Preacher
			require.NoError(t, testDB.First(&p, "id = ?", "p-1").Error)
			assert.Equal(t, "Updated Preacher", p.Name)

			var tag entity.Tag
			require.NoError(t, testDB.First(&tag, "id = ?", "t-1").Error)
			assert.Equal(t, "updated-tag", tag.Name)
		},
	)
}

func TestSyncUsecase_Push_SermonNoteConflictsAndMissingContent(t *testing.T) {
	testDB, u := setupSyncTestEnv(t)
	ctx := context.Background()

	t.Run(
		"appends to Conflicts when server has higher Version with different ContentHash without overwriting",
		func(t *testing.T) {
			note := &entity.SermonNote{
				ID:          "note-conflict-1",
				Title:       "Server Title",
				Version:     5,
				ContentHash: "hash-server-5",
				PlainText:   "Server content",
			}
			require.NoError(t, testDB.Create(note).Error)

			batch := &model.SyncBatch{
				Notes: []*entity.SermonNote{
					{
						ID:          "note-conflict-1",
						Title:       "Client Overwrite Attempt",
						Version:     3,
						ContentHash: "hash-client-3",
						PlainText:   "Client content",
					},
				},
			}

			result, err := u.Push(ctx, batch)
			require.NoError(t, err)
			assert.Contains(t, result.Conflicts, "note-conflict-1")
			assert.NotContains(t, result.MissingContentIDs, "note-conflict-1")

			// Verify record on server was NOT overwritten
			var stored entity.SermonNote
			require.NoError(
				t,
				testDB.First(&stored, "id = ?", "note-conflict-1").Error,
			)
			assert.Equal(t, "Server Title", stored.Title)
			assert.Equal(t, 5, stored.Version)
			assert.Equal(t, "hash-server-5", stored.ContentHash)
		},
	)

	t.Run(
		"appends to MissingContentIDs when note is new and saves note with children",
		func(t *testing.T) {
			verseEnd := 4
			batch := &model.SyncBatch{
				Notes: []*entity.SermonNote{
					{
						ID:          "note-new-1",
						Title:       "Brand New Note",
						Version:     1,
						ContentHash: "hash-new-1",
						PlainText:   "New note text",
						Verses: []entity.NoteVerse{
							{
								Translation: "ESV",
								BookID:      1,
								Chapter:     1,
								VerseStart:  1,
								VerseEnd:    &verseEnd,
								DisplayRef:  "Gen 1:1-4",
							},
						},
						Tags: []entity.Tag{
							{
								ID:   "tag-new-1",
								Name: "Creation",
							},
						},
					},
				},
			}

			result, err := u.Push(ctx, batch)
			require.NoError(t, err)
			assert.Contains(t, result.MissingContentIDs, "note-new-1")
			assert.Empty(t, result.Conflicts)

			// Verify note and children were created
			var stored entity.SermonNote
			require.NoError(
				t,
				testDB.First(&stored, "id = ?", "note-new-1").Error,
			)
			assert.Equal(t, "Brand New Note", stored.Title)
			assert.Equal(t, "hash-new-1", stored.ContentHash)

			var verses []entity.NoteVerse
			require.NoError(
				t,
				testDB.Find(&verses, "note_id = ?", "note-new-1").Error,
			)
			require.Len(t, verses, 1)
			assert.Equal(t, "ESV", verses[0].Translation)

			var fts entity.NotesFTS
			require.NoError(
				t,
				testDB.First(&fts, "note_id = ?", "note-new-1").Error,
			)
			assert.Equal(t, "Brand New Note", fts.Title)
		},
	)

	t.Run(
		"appends to MissingContentIDs when existing note has different ContentHash and updates note",
		func(t *testing.T) {
			note := &entity.SermonNote{
				ID:          "note-update-diff-hash",
				Title:       "Original Title",
				Version:     2,
				ContentHash: "hash-v1",
				PlainText:   "Text v1",
			}
			require.NoError(t, testDB.Create(note).Error)

			batch := &model.SyncBatch{
				Notes: []*entity.SermonNote{
					{
						ID:          "note-update-diff-hash",
						Title:       "Updated Title",
						Version:     3,
						ContentHash: "hash-v2",
						PlainText:   "Text v2",
					},
				},
			}

			result, err := u.Push(ctx, batch)
			require.NoError(t, err)
			assert.Contains(
				t,
				result.MissingContentIDs,
				"note-update-diff-hash",
			)
			assert.Empty(t, result.Conflicts)

			var stored entity.SermonNote
			require.NoError(
				t,
				testDB.First(&stored, "id = ?", "note-update-diff-hash").Error,
			)
			assert.Equal(t, "Updated Title", stored.Title)
			assert.Equal(t, "hash-v2", stored.ContentHash)
			assert.Equal(t, 3, stored.Version)
		},
	)

	t.Run(
		"does NOT append to MissingContentIDs when existing note has same ContentHash",
		func(t *testing.T) {
			note := &entity.SermonNote{
				ID:          "note-same-hash",
				Title:       "Original Title",
				Version:     2,
				ContentHash: "hash-same-content",
				PlainText:   "Same text",
			}
			require.NoError(t, testDB.Create(note).Error)

			batch := &model.SyncBatch{
				Notes: []*entity.SermonNote{
					{
						ID:          "note-same-hash",
						Title:       "Metadata Only Change",
						Version:     3,
						ContentHash: "hash-same-content",
						PlainText:   "Same text",
					},
				},
			}

			result, err := u.Push(ctx, batch)
			require.NoError(t, err)
			assert.NotContains(t, result.MissingContentIDs, "note-same-hash")
			assert.Empty(t, result.Conflicts)

			var stored entity.SermonNote
			require.NoError(
				t,
				testDB.First(&stored, "id = ?", "note-same-hash").Error,
			)
			assert.Equal(t, "Metadata Only Change", stored.Title)
			assert.Equal(t, 3, stored.Version)
		},
	)
}

func TestSyncUsecase_Pull(t *testing.T) {
	testDB, u := setupSyncTestEnv(t)
	ctx := context.Background()

	now := time.Now()
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
			res, err := u.Pull(ctx, now.UnixMilli())
			require.NoError(t, err)
			require.NotNil(t, res)

			assert.Greater(t, res.ServerTime, int64(0))

			var gIDs []string
			for _, g := range res.Gatherings {
				gIDs = append(gIDs, g.ID)
			}
			assert.Contains(t, gIDs, "g-pull-act")
			assert.Contains(t, gIDs, "g-pull-del")
			assert.NotContains(t, gIDs, "g-pull-old")

			require.Len(t, res.Series, 1)
			assert.Equal(t, "s-pull-act", res.Series[0].ID)

			require.Len(t, res.Preachers, 1)
			assert.Equal(t, "p-pull-act", res.Preachers[0].ID)

			require.Len(t, res.Tags, 1)
			assert.Equal(t, "t-pull-act", res.Tags[0].ID)

			var noteIDs []string
			for _, n := range res.Notes {
				noteIDs = append(noteIDs, n.ID)
			}
			assert.Contains(t, noteIDs, "n-pull-act")
			assert.Contains(t, noteIDs, "n-pull-del")
		},
	)

	t.Run("since=0 returns all records", func(t *testing.T) {
		res, err := u.Pull(ctx, 0)
		require.NoError(t, err)
		require.NotNil(t, res)

		var gIDs []string
		for _, g := range res.Gatherings {
			gIDs = append(gIDs, g.ID)
		}
		assert.Contains(t, gIDs, "g-pull-old")
	})
}
