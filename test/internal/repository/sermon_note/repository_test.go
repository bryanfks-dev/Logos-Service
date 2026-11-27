package sermonnote_test

import (
	"context"
	"fmt"
	"testing"
	"time"
	"uuid"

	"github.com/bryanfks-dev/logos/infra/fs"
	sermonnote "github.com/bryanfks-dev/logos/internal/repository/sermon_note"
	tagrepo "github.com/bryanfks-dev/logos/internal/repository/tag"
	"github.com/bryanfks-dev/logos/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSermonNoteTestDB(t *testing.T) *gorm.DB {
	dbName := fmt.Sprintf(
		"file:sermon_note_repo_%s?mode=memory&cache=shared",
		uuid.NewV7().String(),
	)
	testDB, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	require.NoError(t, err)

	err = testDB.AutoMigrate(
		&entity.Gathering{},
		&entity.SermonSeries{},
		&entity.Preacher{},
		&entity.Tag{},
		&entity.SermonNote{},
		&entity.SermonNoteAttachment{},
		&entity.NoteVerse{},
		&entity.NoteTag{},
		&entity.NotesFTS{},
	)
	require.NoError(t, err)

	return testDB
}

func TestSermonNoteRepository(t *testing.T) {
	ctx := context.Background()
	testDB := setupSermonNoteTestDB(t)
	testLog := zap.NewNop()

	repo := sermonnote.NewSermonNoteRepository(testDB, fs.NewManager(), testLog)
	tRepo := tagrepo.NewTagRepository(testDB, testLog)

	graceTag := &entity.Tag{Name: "Faith"}
	assert.NoError(t, tRepo.Create(ctx, graceTag))

	now := time.Now().UTC()

	note := &entity.SermonNote{
		Title:       "Faith and Works",
		PlainText:   "Faith without works is dead",
		ContentJSON: `{"root":{"type":"doc","children":[{"type":"paragraph","text":"Faith without works"}]}}`,
		SermonDate:  now,
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

	scriptureQuery := entity.ScriptureQuery{
		Translation: "ESV",
		BookID:      59,
		Chapter:     2,
		VerseStart:  14,
	}
	byScripture, err := repo.FindByScripture(ctx, scriptureQuery)
	assert.NoError(t, err)
	assert.Len(t, byScripture, 1)

	tagFilter := entity.SermonNoteFilter{
		TagID: &graceTag.ID,
	}
	filtered, err := repo.FindAll(ctx, tagFilter)
	assert.NoError(t, err)
	assert.Len(t, filtered, 1)

	assert.NoError(t, repo.Delete(ctx, note.ID))

	emptyNoteFilter := entity.SermonNoteFilter{}
	remaining, err := repo.FindAll(ctx, emptyNoteFilter)
	assert.NoError(t, err)
	assert.Empty(t, remaining)
}

func TestSermonNoteRepository_FindUpdatedSinceAndBatch(t *testing.T) {
	ctx := context.Background()
	testDB := setupSermonNoteTestDB(t)
	testLog := zap.NewNop()
	repo := sermonnote.NewSermonNoteRepository(testDB, fs.NewManager(), testLog)

	now := time.Now().UTC()
	future := now.Add(2 * time.Hour)

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

	require.NoError(t, repo.CreateBatch(ctx, nil))
	require.NoError(t, repo.UpdateBatch(ctx, nil))

	items := []*entity.SermonNote{
		{
			ID:        "nb-1",
			Title:     "Batch Note 1",
			PlainText: "Body 1",
			Verses: []entity.NoteVerse{
				{
					Translation: "ESV",
					BookID:      1,
					Chapter:     1,
					VerseStart:  1,
					DisplayRef:  "Gen 1:1",
				},
			},
		},
		{
			ID:        "nb-2",
			Title:     "Batch Note 2",
			PlainText: "Body 2",
		},
	}
	require.NoError(t, repo.CreateBatch(ctx, items))

	items[0].Title = "Batch Note 1 Updated"
	items[1].Title = "Batch Note 2 Updated"
	require.NoError(t, repo.UpdateBatch(ctx, items))

	fetched, err := repo.FindByIDsUnscoped(ctx, []string{"nb-1", "nb-2"})
	require.NoError(t, err)
	require.Len(t, fetched, 2)
}
