package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/bryanfks-dev/selah/infrastructure/db"
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

func setupSyncTestEnv(t *testing.T) (*gorm.DB, usecase.SyncUsecase) {
	testDB, err := gorm.Open(
		sqlite.Open("file::memory:?cache=shared"),
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
	noteRepo := repository.NewSermonNoteRepository(testDB, log)
	tagRepo := repository.NewTagRepository(testDB, log)
	syncOutboxRepo := repository.NewSyncOutboxRepository(testDB, log)

	syncUsecase := usecase.NewSyncUsecase(
		txManager,
		noteRepo,
		tagRepo,
		syncOutboxRepo,
		log,
	)

	return testDB, syncUsecase
}

func TestSyncNotes(t *testing.T) {
	ctx := context.Background()
	_, syncUsecase := setupSyncTestEnv(t)

	t.Run("creates new note pushed by client", func(t *testing.T) {
		req := &model.SyncNotesRequest{
			Notes: []model.SyncNoteItem{
				{
					ID:          "sync-note-1",
					Title:       "Grace and Peace",
					PlainText:   "Grace to you and peace",
					ContentJSON: `{"text":"Grace"}`,
					Version:     1,
					Tags:        []string{"Grace", "Peace"},
					Verses: []model.SyncNoteVerse{
						{
							Translation: "ESV",
							BookID:      1,
							Chapter:     1,
							VerseStart:  1,
							DisplayRef:  "Gen 1:1",
						},
					},
				},
			},
		}

		res, err := syncUsecase.SyncNotes(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Contains(t, res.AppliedIDs, "sync-note-1")
		assert.False(t, res.SyncedAt.IsZero())

		// Verify note is pulled back
		var found bool
		for _, n := range res.Notes {
			if n.ID == "sync-note-1" {
				found = true
				assert.Equal(t, "Grace and Peace", n.Title)
				assert.ElementsMatch(t, []string{"Grace", "Peace"}, n.Tags)
			}
		}
		assert.True(t, found)
	})

	t.Run("client pushes deleted note that does not exist", func(t *testing.T) {
		req := &model.SyncNotesRequest{
			Notes: []model.SyncNoteItem{
				{
					ID:        "non-existent-del-note",
					Title:     "Deleted note",
					IsDeleted: true,
				},
			},
		}

		res, err := syncUsecase.SyncNotes(ctx, req)
		assert.NoError(t, err)
		assert.Contains(t, res.AppliedIDs, "non-existent-del-note")
	})

	t.Run(
		"updates existing note with equal or higher version",
		func(t *testing.T) {
			req := &model.SyncNotesRequest{
				Notes: []model.SyncNoteItem{
					{
						ID:          "sync-note-1",
						Title:       "Grace and Peace (Updated)",
						PlainText:   "Updated content",
						ContentJSON: `{"text":"Updated"}`,
						Version:     1, // existing was 1, so version >= existing
					},
				},
			}

			res, err := syncUsecase.SyncNotes(ctx, req)
			assert.NoError(t, err)
			assert.Contains(t, res.AppliedIDs, "sync-note-1")

			var found bool
			for _, n := range res.Notes {
				if n.ID == "sync-note-1" {
					found = true
					assert.Equal(t, "Grace and Peace (Updated)", n.Title)
					assert.Equal(t, 2, n.Version)
				}
			}
			assert.True(t, found)
		},
	)

	t.Run(
		"conflict resolution when client version is lower than server version",
		func(t *testing.T) {
			// Existing note version is now 2
			req := &model.SyncNotesRequest{
				Notes: []model.SyncNoteItem{
					{
						ID:        "sync-note-1",
						Title:     "Old Client Version Trying To Overwrite",
						Version:   1, // lower than server version 2
						IsDeleted: false,
					},
				},
			}

			res, err := syncUsecase.SyncNotes(ctx, req)
			assert.NoError(t, err)
			// Should NOT be in applied IDs
			assert.NotContains(t, res.AppliedIDs, "sync-note-1")

			// Server version should be in response
			var found bool
			for _, n := range res.Notes {
				if n.ID == "sync-note-1" {
					found = true
					assert.Equal(t, "Grace and Peace (Updated)", n.Title)
				}
			}
			assert.True(t, found)
		},
	)

	t.Run(
		"deletes existing note when client sends is_deleted",
		func(t *testing.T) {
			req := &model.SyncNotesRequest{
				Notes: []model.SyncNoteItem{
					{
						ID:        "sync-note-1",
						Version:   2,
						IsDeleted: true,
					},
				},
			}

			res, err := syncUsecase.SyncNotes(ctx, req)
			assert.NoError(t, err)
			assert.Contains(t, res.AppliedIDs, "sync-note-1")

			var found bool
			for _, n := range res.Notes {
				if n.ID == "sync-note-1" {
					found = true
					assert.True(t, n.IsDeleted)
				}
			}
			assert.True(t, found)
		},
	)

	t.Run("handles nil request gracefully", func(t *testing.T) {
		res, err := syncUsecase.SyncNotes(ctx, nil)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("pulls notes updated since last_synced_at", func(t *testing.T) {
		// Create a fresh note
		reqCreate := &model.SyncNotesRequest{
			Notes: []model.SyncNoteItem{
				{
					ID:      "note-recent",
					Title:   "Recent Note",
					Version: 1,
				},
			},
		}
		resCreate, err := syncUsecase.SyncNotes(ctx, reqCreate)
		require.NoError(t, err)

		// Sync with last_synced_at before creation time
		pastTime := resCreate.SyncedAt.Add(-1 * time.Hour)
		resSync, err := syncUsecase.SyncNotes(ctx, &model.SyncNotesRequest{
			LastSyncedAt: &pastTime,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, resSync.Notes)
	})
}
