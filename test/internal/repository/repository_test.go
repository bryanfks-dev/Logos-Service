package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/bryanfks-dev/selah/db/correlation"
	"github.com/bryanfks-dev/selah/internal/repository"
	"github.com/bryanfks-dev/selah/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
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
		&entity.BibleVerse{},
		&entity.BibleFTS{},
		&entity.SyncOutboxEntry{},
	)
	require.NoError(t, err)

	return testDB
}

func TestGatheringRepository(t *testing.T) {
	ctx := context.Background()

	testDB := setupTestDB(t)

	testLog := zap.NewNop()

	repo := repository.NewGatheringRepository(testDB, testLog)

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

func TestSermonSeriesRepository(t *testing.T) {
	ctx := context.Background()

	testDB := setupTestDB(t)

	testLog := zap.NewNop()

	repo := repository.NewSermonSeriesRepository(testDB, testLog)

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

func TestPreacherRepository(t *testing.T) {
	ctx := context.Background()

	testDB := setupTestDB(t)

	testLog := zap.NewNop()

	repo := repository.NewPreacherRepository(testDB, testLog)

	p := &entity.Preacher{Name: "John Doe", ChurchName: "Grace Church"}

	assert.NoError(t, repo.Create(ctx, p))

	p.ChurchName = "Grace Community Church"

	assert.NoError(t, repo.Update(ctx, p))

	fetched, err := repo.FindByID(ctx, p.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Grace Community Church", fetched.ChurchName)

	p2 := &entity.Preacher{Name: "Jane Smith"}

	assert.NoError(t, repo.Create(ctx, p2))

	filtered, err := repo.FindAll(ctx, entity.PreacherFilter{NamePrefix: "Jo"})
	assert.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "John Doe", filtered[0].Name)

	assert.NoError(t, repo.Delete(ctx, p.ID))

	filteredAfterDelete, err := repo.FindAll(ctx, entity.PreacherFilter{})
	assert.NoError(t, err)
	assert.Len(t, filteredAfterDelete, 1)
}

func TestTagRepository(t *testing.T) {
	ctx := context.Background()

	testDB := setupTestDB(t)

	testLog := zap.NewNop()

	repo := repository.NewTagRepository(testDB, testLog)

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

func TestSermonNoteRepository(t *testing.T) {
	ctx := context.Background()

	testDB := setupTestDB(t)

	testLog := zap.NewNop()

	repo := repository.NewSermonNoteRepository(testDB, testLog)

	tagRepo := repository.NewTagRepository(testDB, testLog)

	graceTag := &entity.Tag{Name: "Faith"}

	assert.NoError(t, tagRepo.Create(ctx, graceTag))

	now := time.Now()

	note := &entity.SermonNote{
		Title:       "Faith and Works",
		PlainText:   "Faith without works is dead",
		ContentJSON: `{"root":{"type":"doc","children":[{"type":"paragraph","text":"Faith without works"}]}}`,
		SermonDate:  &now,
		Verses: []entity.NoteVerse{
			{
				Translation: "ESV",
				BookID:      59,
				Chapter:     2,
				VerseStart:  14,
				DisplayRef:  "James 2:14",
			},
		},
		Tags: []entity.Tag{
			*graceTag,
		},
		FTS: &entity.NotesFTS{
			Title:        "Faith and Works",
			PreacherName: "Pastor Tim",
			PlainText:    "Faith without works is dead",
		},
	}

	assert.NoError(t, repo.Create(ctx, note))
	assert.NotEmpty(t, note.ID)

	fetched, err := repo.FindByID(ctx, note.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Faith and Works", fetched.Title)
	assert.Equal(t, note.ContentJSON, fetched.ContentJSON)
	assert.Len(t, fetched.Verses, 1)
	assert.Len(t, fetched.Tags, 1)
	assert.NotNil(t, fetched.FTS)

	note.Title = "Faith and Works (Updated)"

	assert.NoError(t, repo.Update(ctx, note))

	searched, err := repo.Search(ctx, "Pastor Tim")
	assert.NoError(t, err)
	assert.Len(t, searched, 1)

	byScripture, err := repo.FindByScripture(ctx, entity.ScriptureQuery{
		Translation: "ESV",
		BookID:      59,
		Chapter:     2,
		VerseStart:  14,
	})
	assert.NoError(t, err)
	assert.Len(t, byScripture, 1)

	filtered, err := repo.FindAll(ctx, entity.SermonNoteFilter{
		TagID: &graceTag.ID,
	})
	assert.NoError(t, err)
	assert.Len(t, filtered, 1)

	assert.NoError(t, repo.Delete(ctx, note.ID))

	remaining, err := repo.FindAll(ctx, entity.SermonNoteFilter{})
	assert.NoError(t, err)
	assert.Empty(t, remaining)
}

func TestBibleRepository(t *testing.T) {
	ctx := context.Background()

	testDB := setupTestDB(t)

	testLog := zap.NewNop()

	repo := repository.NewBibleRepository(testDB, testLog)

	verses := []entity.BibleVerse{
		{
			Translation: "ESV",
			BookID:      1,
			BookName:    "Genesis",
			Chapter:     1,
			Verse:       1,
			Text:        "In the beginning, God created the heavens and the earth.",
		},
		{
			Translation: "ESV",
			BookID:      1,
			BookName:    "Genesis",
			Chapter:     1,
			Verse:       2,
			Text:        "The earth was without form and void.",
		},
		{
			Translation: "NIV",
			BookID:      1,
			BookName:    "Genesis",
			Chapter:     1,
			Verse:       1,
			Text:        "In the beginning God created the heavens and the earth.",
		},
	}

	assert.NoError(t, repo.ImportTranslationPack(ctx, verses))

	translations, err := repo.ListAvailableTranslations(ctx)
	assert.NoError(t, err)
	assert.Len(t, translations, 2)

	end := 2

	fetched, err := repo.GetVerses(ctx, entity.VerseQuery{
		Translation: "ESV",
		BookID:      1,
		Chapter:     1,
		VerseStart:  1,
		VerseEnd:    &end,
	})
	assert.NoError(t, err)
	assert.Len(t, fetched, 2)

	searched, err := repo.SearchByText(ctx, "ESV", "beginning")
	assert.NoError(t, err)
	assert.Len(t, searched, 1)
}

func TestSyncOutboxRepository(t *testing.T) {
	ctx := context.Background()

	testDB := setupTestDB(t)

	testLog := zap.NewNop()

	repo := repository.NewSyncOutboxRepository(testDB, testLog)

	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	entry1 := &entity.SyncOutboxEntry{
		EntityType: "sermon_notes",
		EntityID:   "note-1",
		Operation:  "create",
		QueuedAt:   baseTime.Add(1 * time.Second),
	}

	entry2 := &entity.SyncOutboxEntry{
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

	assert.NoError(t, repo.MarkSynced(ctx, []string{entry1.ID}))

	remaining, err := repo.GetPendingBatch(ctx, 10)
	assert.NoError(t, err)
	assert.Len(t, remaining, 1)

	assert.NoError(t, repo.ClearSynced(ctx, baseTime.Add(3*time.Second)))

	cleared, err := repo.GetPendingBatch(ctx, 10)
	assert.NoError(t, err)
	assert.Empty(t, cleared)
}

func TestTransactionPropagation(t *testing.T) {
	ctx := context.Background()

	testDB := setupTestDB(t)

	testLog := zap.NewNop()

	repo := repository.NewGatheringRepository(testDB, testLog)

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

func TestRepositories_FindUpdatedSince(t *testing.T) {
	ctx := context.Background()
	testDB := setupTestDB(t)
	testLog := zap.NewNop()

	now := time.Now()
	past := now.Add(-2 * time.Hour)
	future := now.Add(2 * time.Hour)

	t.Run("GatheringRepository includes soft-deleted rows", func(t *testing.T) {
		repo := repository.NewGatheringRepository(testDB, testLog)
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
	})

	t.Run(
		"SermonSeriesRepository includes soft-deleted rows",
		func(t *testing.T) {
			repo := repository.NewSermonSeriesRepository(testDB, testLog)
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
		},
	)

	t.Run("PreacherRepository includes soft-deleted rows", func(t *testing.T) {
		repo := repository.NewPreacherRepository(testDB, testLog)
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
	})

	t.Run("TagRepository returns tags updated since", func(t *testing.T) {
		repo := repository.NewTagRepository(testDB, testLog)
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
	})

	t.Run(
		"SermonNoteRepository includes soft-deleted rows and hydrates notes",
		func(t *testing.T) {
			repo := repository.NewSermonNoteRepository(testDB, testLog)
			nActive := &entity.SermonNote{
				ID:        "n-act",
				Title:     "Active Note",
				UpdatedAt: future,
			}
			nDeleted := &entity.SermonNote{
				ID:        "n-del",
				Title:     "Deleted Note",
				UpdatedAt: future,
				DeletedAt: gorm.DeletedAt{Time: future, Valid: true},
			}
			require.NoError(t, testDB.Create(nActive).Error)
			require.NoError(t, testDB.Create(nDeleted).Error)

			results, err := repo.FindUpdatedSince(ctx, now)
			require.NoError(t, err)
			var ids []string
			for _, r := range results {
				ids = append(ids, r.ID)
			}
			assert.Contains(t, ids, "n-act")
			assert.Contains(t, ids, "n-del")
		},
	)
}
