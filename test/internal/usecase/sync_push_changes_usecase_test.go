package usecase_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
	"uuid"

	"github.com/bryanfks-dev/logos/infra/db"
	"github.com/bryanfks-dev/logos/infra/fs"
	"github.com/bryanfks-dev/logos/internal/model"
	"github.com/bryanfks-dev/logos/internal/repository"
	"github.com/bryanfks-dev/logos/internal/usecase"
	"github.com/bryanfks-dev/logos/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSyncTestEnv(
	t *testing.T,
) (*gorm.DB, *usecase.SyncPushChangesUsecase) {
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
	sermonNoteRepository := repository.NewSermonNoteRepository(
		testDB,
		fs.NewManager(),
		log,
	)
	gatheringRepository := repository.NewGatheringRepository(testDB, log)
	sermonSeriesRepository := repository.NewSermonSeriesRepository(testDB, log)
	preacherRepository := repository.NewPreacherRepository(testDB, log)
	tagRepository := repository.NewTagRepository(testDB, log)

	syncPushChangesUsecase := usecase.NewSyncPushChangesUsecase(
		&usecase.SyncPushChangesUsecaseOption{
			TxManager:              txManager,
			SermonNoteRepository:   sermonNoteRepository,
			GatheringRepository:    gatheringRepository,
			SermonSeriesRepository: sermonSeriesRepository,
			PreacherRepository:     preacherRepository,
			TagRepository:          tagRepository,
			Log:                    log,
		},
	)

	return testDB, syncPushChangesUsecase
}

func TestSyncPushChangesUsecase_Push_NilBatch(t *testing.T) {
	_, u := setupSyncTestEnv(t)
	ctx := context.Background()

	result, err := u.Call(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.SyncedAt.IsZero())
	assert.Empty(t, result.Conflicts)
	assert.Empty(t, result.MissingContentIDs)
}

func TestSyncPushChangesUsecase_Push_UpsertMetadata(t *testing.T) {
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
			batch := &model.SyncChangesBatch{
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
				SermonSeries: []*entity.SermonSeries{
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

			result, err := u.Call(ctx, batch)
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
			batch := &model.SyncChangesBatch{
				Gatherings: []*entity.Gathering{
					{
						ID:        "g-1",
						Name:      "Should Not Overwrite Gathering",
						UpdatedAt: olderTime,
					},
				},
				SermonSeries: []*entity.SermonSeries{
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

			result, err := u.Call(ctx, batch)
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

func TestSyncPushChangesUsecase_Push_SermonNoteConflictsAndMissingContent(
	t *testing.T,
) {
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

			batch := &model.SyncChangesBatch{
				SermonNotes: []*entity.SermonNote{
					{
						ID:          "note-conflict-1",
						Title:       "Client Overwrite Attempt",
						Version:     3,
						ContentHash: "hash-client-3",
						PlainText:   "Client content",
					},
				},
			}

			result, err := u.Call(ctx, batch)
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
			batch := &model.SyncChangesBatch{
				SermonNotes: []*entity.SermonNote{
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

			result, err := u.Call(ctx, batch)
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

			batch := &model.SyncChangesBatch{
				SermonNotes: []*entity.SermonNote{
					{
						ID:          "note-update-diff-hash",
						Title:       "Updated Title",
						Version:     3,
						ContentHash: "hash-v2",
						PlainText:   "Text v2",
					},
				},
			}

			result, err := u.Call(ctx, batch)
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

			batch := &model.SyncChangesBatch{
				SermonNotes: []*entity.SermonNote{
					{
						ID:          "note-same-hash",
						Title:       "Metadata Only Change",
						Version:     3,
						ContentHash: "hash-same-content",
						PlainText:   "Same text",
					},
				},
			}

			result, err := u.Call(ctx, batch)
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

type failingSermonNoteRepo struct {
	repository.SermonNoteRepository
	findUpdatedSinceErr error
	findByIDErr         error
	createErr           error
	updateErr           error
}

func (f *failingSermonNoteRepo) FindUpdatedSince(
	ctx context.Context,
	since time.Time,
) ([]*entity.SermonNote, error) {
	if f.findUpdatedSinceErr != nil {
		return nil, f.findUpdatedSinceErr
	}

	return f.SermonNoteRepository.FindUpdatedSince(ctx, since)
}

func (f *failingSermonNoteRepo) FindByID(
	ctx context.Context,
	id string,
) (*entity.SermonNote, error) {
	if f.findByIDErr != nil {
		return nil, f.findByIDErr
	}

	return f.SermonNoteRepository.FindByID(ctx, id)
}

func (f *failingSermonNoteRepo) FindByIDsUnscoped(
	ctx context.Context,
	ids []string,
) ([]*entity.SermonNote, error) {
	if f.findByIDErr != nil {
		return nil, f.findByIDErr
	}

	return f.SermonNoteRepository.FindByIDsUnscoped(ctx, ids)
}

func (f *failingSermonNoteRepo) Create(
	ctx context.Context,
	note *entity.SermonNote,
) error {
	if f.createErr != nil {
		return f.createErr
	}

	return f.SermonNoteRepository.Create(ctx, note)
}

func (f *failingSermonNoteRepo) Update(
	ctx context.Context,
	note *entity.SermonNote,
) error {
	if f.updateErr != nil {
		return f.updateErr
	}

	return f.SermonNoteRepository.Update(ctx, note)
}

type failingGatheringRepo struct {
	repository.GatheringRepository
	findUpdatedSinceErr error
	findByIDErr         error
}

func (f *failingGatheringRepo) FindUpdatedSince(
	ctx context.Context,
	since time.Time,
) ([]*entity.Gathering, error) {
	if f.findUpdatedSinceErr != nil {
		return nil, f.findUpdatedSinceErr
	}

	return f.GatheringRepository.FindUpdatedSince(ctx, since)
}

func (f *failingGatheringRepo) FindByID(
	ctx context.Context,
	id string,
) (*entity.Gathering, error) {
	if f.findByIDErr != nil {
		return nil, f.findByIDErr
	}

	return f.GatheringRepository.FindByID(ctx, id)
}

func (f *failingGatheringRepo) FindByIDs(
	ctx context.Context,
	ids []string,
) ([]*entity.Gathering, error) {
	if f.findByIDErr != nil {
		return nil, f.findByIDErr
	}

	return f.GatheringRepository.FindByIDs(ctx, ids)
}

func TestSyncPushChangesUsecase_Push_EdgeCasesAndFileHash(t *testing.T) {
	testDB, u := setupSyncTestEnv(t)
	ctx := context.Background()

	t.Run(
		"skips nil items, generates UUIDv7 IDs, and supports alias slices",
		func(t *testing.T) {
			g := &entity.Gathering{Name: "Auto ID Gathering"}
			s := &entity.SermonSeries{Name: "Auto ID Series"}
			p := &entity.Preacher{Name: "Auto ID Preacher"}
			tag := &entity.Tag{Name: "auto-id-tag"}
			note := &entity.SermonNote{
				Title:       "Auto ID Note",
				Version:     1,
				ContentHash: "hash-auto-id",
			}

			batch := &model.SyncChangesBatch{
				Gatherings:   []*entity.Gathering{nil, g},
				SermonSeries: []*entity.SermonSeries{nil, s},
				Preachers:    []*entity.Preacher{nil, p},
				Tags:         []*entity.Tag{nil, tag},
				SermonNotes:  []*entity.SermonNote{nil, note},
			}

			res, err := u.Call(ctx, batch)
			require.NoError(t, err)
			require.NotNil(t, res)

			for _, id := range []string{g.ID, s.ID, p.ID, tag.ID, note.ID} {
				parsed, parseErr := uuid.Parse(id)
				require.NoError(t, parseErr)
				assert.Equal(t, byte(7), parsed[6]>>4)
			}

			assert.Contains(t, res.MissingContentIDs, note.ID)
		},
	)

	t.Run(
		"computes content hash from hydrated note file when present",
		func(t *testing.T) {
			content := `{"ops":[{"insert":"hello"}]}`
			sum := sha256.Sum256([]byte(content))
			contentHash := hex.EncodeToString(sum[:])

			noteFile := filepath.Join(t.TempDir(), "hydrated-note.json")
			require.NoError(t, os.WriteFile(noteFile, []byte(content), 0o644))

			existing := &entity.SermonNote{
				ID:          "note-hydrated-file",
				Title:       "Hydrated",
				Version:     1,
				FilePath:    noteFile,
				ContentHash: "stale-db-hash",
			}
			require.NoError(t, testDB.Create(existing).Error)

			// Matching file hash -> not missing content
			matchingBatch := &model.SyncChangesBatch{
				SermonNotes: []*entity.SermonNote{
					{
						ID:          "note-hydrated-file",
						Title:       "Hydrated Updated",
						Version:     2,
						FilePath:    noteFile,
						ContentHash: contentHash,
						ContentJSON: content,
					},
				},
			}
			res, err := u.Call(ctx, matchingBatch)
			require.NoError(t, err)
			assert.NotContains(t, res.MissingContentIDs, "note-hydrated-file")

			// Differing hash -> reported as missing content
			diffBatch := &model.SyncChangesBatch{
				SermonNotes: []*entity.SermonNote{
					{
						ID:          "note-hydrated-file",
						Title:       "Hydrated Updated Again",
						Version:     3,
						FilePath:    noteFile,
						ContentHash: "different-client-hash",
					},
				},
			}
			res, err = u.Call(ctx, diffBatch)
			require.NoError(t, err)
			assert.Contains(t, res.MissingContentIDs, "note-hydrated-file")
		},
	)
}

func TestSyncPushChangesUsecase_ErrorPropagationAndRollback(t *testing.T) {
	ctx := context.Background()

	t.Run(
		"Call rolls back transaction when repository fails",
		func(t *testing.T) {
			testDB, _ := setupSyncTestEnv(t)
			log := zap.NewNop()
			expectedErr := errors.New("note create failed")

			u := usecase.NewSyncPushChangesUsecase(
				&usecase.SyncPushChangesUsecaseOption{
					TxManager: db.NewTxManager(testDB),
					SermonNoteRepository: &failingSermonNoteRepo{
						SermonNoteRepository: repository.NewSermonNoteRepository(
							testDB,
							fs.NewManager(),
							log,
						),
						createErr: expectedErr,
					},
					GatheringRepository: repository.NewGatheringRepository(
						testDB,
						log,
					),
					SermonSeriesRepository: repository.NewSermonSeriesRepository(
						testDB,
						log,
					),
					PreacherRepository: repository.NewPreacherRepository(
						testDB,
						log,
					),
					TagRepository: repository.NewTagRepository(testDB, log),
					Log:           log,
				},
			)

			batch := &model.SyncChangesBatch{
				Gatherings: []*entity.Gathering{
					{ID: "g-rollback", Name: "Should Roll Back"},
				},
				SermonNotes: []*entity.SermonNote{
					{ID: "n-fail", Title: "Fail"},
				},
			}

			res, err := u.Call(ctx, batch)
			assert.Nil(t, res)
			assert.ErrorIs(t, err, expectedErr)

			var count int64
			require.NoError(
				t,
				testDB.Model(&entity.Gathering{}).
					Where("id = ?", "g-rollback").
					Count(&count).
					Error,
			)
			assert.Equal(t, int64(0), count)
		},
	)
}

func TestSyncPushChangesUsecase_Push_MultiClientScenarios(t *testing.T) {
	t.Run(
		"detects conflict when two clients push same version with different ContentHash",
		func(t *testing.T) {
			testDB, u := setupSyncTestEnv(t)
			ctx := context.Background()

			clientABatch := &model.SyncChangesBatch{
				SermonNotes: []*entity.SermonNote{
					{
						ID:          "note-same-ver",
						Title:       "Client A Title",
						Version:     2,
						ContentHash: "hash-client-a",
						PlainText:   "Client A content",
					},
				},
			}
			resA, err := u.Call(ctx, clientABatch)
			require.NoError(t, err)
			assert.Empty(t, resA.Conflicts)

			clientBBatch := &model.SyncChangesBatch{
				SermonNotes: []*entity.SermonNote{
					{
						ID:          "note-same-ver",
						Title:       "Client B Title",
						Version:     2,
						ContentHash: "hash-client-b",
						PlainText:   "Client B content",
					},
				},
			}
			resB, err := u.Call(ctx, clientBBatch)
			require.NoError(t, err)
			assert.Contains(t, resB.Conflicts, "note-same-ver")

			var stored entity.SermonNote
			require.NoError(
				t,
				testDB.First(&stored, "id = ?", "note-same-ver").Error,
			)
			assert.Equal(t, "Client A Title", stored.Title)
			assert.Equal(t, "hash-client-a", stored.ContentHash)
		},
	)

	t.Run(
		"allows idempotent retry when same version and same ContentHash are pushed",
		func(t *testing.T) {
			_, u := setupSyncTestEnv(t)
			ctx := context.Background()

			batch := &model.SyncChangesBatch{
				SermonNotes: []*entity.SermonNote{
					{
						ID:          "note-idempotent",
						Title:       "Initial Title",
						Version:     2,
						ContentHash: "hash-same",
						PlainText:   "Same content",
					},
				},
			}

			res1, err := u.Call(ctx, batch)
			require.NoError(t, err)
			assert.Empty(t, res1.Conflicts)

			res2, err := u.Call(ctx, batch)
			require.NoError(t, err)
			assert.Empty(t, res2.Conflicts)
		},
	)

	t.Run(
		"handles push for a soft-deleted note without primary key collision",
		func(t *testing.T) {
			testDB, u := setupSyncTestEnv(t)
			ctx := context.Background()

			note := &entity.SermonNote{
				ID:          "note-soft-deleted",
				Title:       "Deleted Note",
				Version:     1,
				ContentHash: "hash-v1",
				PlainText:   "Old text",
			}
			require.NoError(t, testDB.Create(note).Error)
			require.NoError(t, testDB.Delete(note).Error)

			batch := &model.SyncChangesBatch{
				SermonNotes: []*entity.SermonNote{
					{
						ID:          "note-soft-deleted",
						Title:       "Restored By Client B",
						Version:     2,
						ContentHash: "hash-v2",
						PlainText:   "New text",
					},
				},
			}

			res, err := u.Call(ctx, batch)
			require.NoError(t, err)
			assert.Empty(t, res.Conflicts)

			var stored entity.SermonNote
			require.NoError(
				t,
				testDB.First(&stored, "id = ?", "note-soft-deleted").Error,
			)
			assert.Equal(t, "Restored By Client B", stored.Title)
			assert.Equal(t, 2, stored.Version)
		},
	)

	t.Run(
		"handles concurrent pushes from two clients for the same note ID",
		func(t *testing.T) {
			testDB, u := setupSyncTestEnv(t)
			ctx := context.Background()

			var (
				wg   sync.WaitGroup
				errA error
				errB error
				resA *model.SyncPushChangesResult
				resB *model.SyncPushChangesResult
			)

			wg.Add(2)
			go func() {
				defer wg.Done()
				batchA := &model.SyncChangesBatch{
					SermonNotes: []*entity.SermonNote{
						{
							ID:          "note-concurrent-1",
							Title:       "Client A Concurrent",
							Version:     1,
							ContentHash: "hash-a",
							PlainText:   "Content A",
						},
					},
				}
				resA, errA = u.Call(ctx, batchA)
			}()

			go func() {
				defer wg.Done()
				batchB := &model.SyncChangesBatch{
					SermonNotes: []*entity.SermonNote{
						{
							ID:          "note-concurrent-1",
							Title:       "Client B Concurrent",
							Version:     1,
							ContentHash: "hash-b",
							PlainText:   "Content B",
						},
					},
				}
				resB, errB = u.Call(ctx, batchB)
			}()

			wg.Wait()

			require.NoError(t, errA)
			require.NoError(t, errB)

			totalConflicts := len(resA.Conflicts) + len(resB.Conflicts)
			assert.Equal(t, 1, totalConflicts)

			var stored entity.SermonNote
			require.NoError(
				t,
				testDB.First(&stored, "id = ?", "note-concurrent-1").Error,
			)
			expectedTitles := []string{
				"Client A Concurrent",
				"Client B Concurrent",
			}
			assert.Contains(t, expectedTitles, stored.Title)
		},
	)

	t.Run(
		"handles duplicate IDs in the same push batch using pre-fetched state",
		func(t *testing.T) {
			testDB, u := setupSyncTestEnv(t)
			ctx := context.Background()

			t1 := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
			t2 := time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC)

			dupBatch := &model.SyncChangesBatch{
				Gatherings: []*entity.Gathering{
					{ID: "g-dup", Name: "First", UpdatedAt: t1},
					{ID: "g-dup", Name: "Second", UpdatedAt: t2},
				},
				SermonNotes: []*entity.SermonNote{
					{
						ID:          "n-dup",
						Title:       "First Note",
						Version:     1,
						ContentHash: "h1",
					},
					{
						ID:          "n-dup",
						Title:       "Second Note",
						Version:     2,
						ContentHash: "h2",
					},
				},
			}
			res, err := u.Call(ctx, dupBatch)
			require.NoError(t, err)
			assert.Empty(t, res.Conflicts)

			var g entity.Gathering
			require.NoError(t, testDB.First(&g, "id = ?", "g-dup").Error)
			assert.Equal(t, "Second", g.Name)

			var n entity.SermonNote
			require.NoError(t, testDB.First(&n, "id = ?", "n-dup").Error)
			assert.Equal(t, "Second Note", n.Title)
			assert.Equal(t, 2, n.Version)
		},
	)

	t.Run(
		"remaps duplicate Tag.Name across clients with different UUIDs and links note_tags",
		func(t *testing.T) {
			testDB, u := setupSyncTestEnv(t)
			ctx := context.Background()

			t1 := time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)
			t2 := time.Date(2026, 2, 1, 11, 0, 0, 0, time.UTC)

			batchA := &model.SyncChangesBatch{
				Tags: []*entity.Tag{
					{ID: "tag-client-a", Name: "Grace", UpdatedAt: t1},
				},
				SermonNotes: []*entity.SermonNote{
					{
						ID:          "note-tag-a",
						Title:       "Note A",
						Version:     1,
						ContentHash: "hash-a",
						Tags: []entity.Tag{
							{ID: "tag-client-a", Name: "Grace"},
						},
					},
				},
			}
			resA, err := u.Call(ctx, batchA)
			require.NoError(t, err)
			assert.Empty(t, resA.Conflicts)

			batchB := &model.SyncChangesBatch{
				Tags: []*entity.Tag{
					{ID: "tag-client-b", Name: "Grace", UpdatedAt: t2},
				},
				SermonNotes: []*entity.SermonNote{
					{
						ID:          "note-tag-b",
						Title:       "Note B",
						Version:     1,
						ContentHash: "hash-b",
						Tags: []entity.Tag{
							{ID: "tag-client-b", Name: "Grace"},
						},
					},
				},
			}
			resB, err := u.Call(ctx, batchB)
			require.NoError(t, err)
			assert.Empty(t, resB.Conflicts)

			var tags []entity.Tag
			require.NoError(t, testDB.Find(&tags, "name = ?", "Grace").Error)
			require.Len(t, tags, 1)
			assert.Equal(t, "tag-client-a", tags[0].ID)

			var noteTags []entity.NoteTag
			require.NoError(
				t,
				testDB.Find(&noteTags, "note_id = ?", "note-tag-b").Error,
			)
			require.Len(t, noteTags, 1)
			assert.Equal(t, "tag-client-a", noteTags[0].TagID)
		},
	)

	t.Run(
		"handles push for soft-deleted gathering without primary key collision",
		func(t *testing.T) {
			testDB, u := setupSyncTestEnv(t)
			ctx := context.Background()

			g := &entity.Gathering{
				ID:   "g-soft-del",
				Name: "Deleted Gathering",
			}
			require.NoError(t, testDB.Create(g).Error)
			require.NoError(t, testDB.Delete(g).Error)

			restoreBatch := &model.SyncChangesBatch{
				Gatherings: []*entity.Gathering{
					{
						ID:        "g-soft-del",
						Name:      "Restored Gathering",
						UpdatedAt: time.Now().UTC(),
					},
				},
			}
			_, err := u.Call(ctx, restoreBatch)
			require.NoError(t, err)

			var stored entity.Gathering
			require.NoError(
				t,
				testDB.First(&stored, "id = ?", "g-soft-del").Error,
			)
			assert.Equal(t, "Restored Gathering", stored.Name)
		},
	)

	t.Run(
		"does not overwrite note file on disk when push transaction rolls back",
		func(t *testing.T) {
			testDB, u := setupSyncTestEnv(t)
			ctx := context.Background()

			noteFile := filepath.Join(t.TempDir(), "rollback-note.json")
			originalContent := `{"ops":[{"insert":"original"}]}`
			origSum := sha256.Sum256([]byte(originalContent))
			origHash := hex.EncodeToString(origSum[:])

			initialBatch := &model.SyncChangesBatch{
				SermonNotes: []*entity.SermonNote{
					{
						ID:          "note-file-rb-1",
						Title:       "Original",
						Version:     1,
						FilePath:    noteFile,
						ContentJSON: originalContent,
						ContentHash: origHash,
					},
				},
			}
			_, err := u.Call(ctx, initialBatch)
			require.NoError(t, err)

			log := zap.NewNop()
			failErr := errors.New("second note failed")
			failOpt := &usecase.SyncPushChangesUsecaseOption{
				TxManager: db.NewTxManager(testDB),
				SermonNoteRepository: &failingSermonNoteRepo{
					SermonNoteRepository: repository.NewSermonNoteRepository(
						testDB,
						fs.NewManager(),
						log,
					),
					createErr: failErr,
				},
				GatheringRepository: repository.NewGatheringRepository(
					testDB,
					log,
				),
				SermonSeriesRepository: repository.NewSermonSeriesRepository(
					testDB,
					log,
				),
				PreacherRepository: repository.NewPreacherRepository(
					testDB,
					log,
				),
				TagRepository: repository.NewTagRepository(testDB, log),
				Log:           log,
			}
			uFail := usecase.NewSyncPushChangesUsecase(failOpt)

			newContent := `{"ops":[{"insert":"should-not-persist"}]}`
			newSum := sha256.Sum256([]byte(newContent))
			newHash := hex.EncodeToString(newSum[:])

			rollbackBatch := &model.SyncChangesBatch{
				SermonNotes: []*entity.SermonNote{
					{
						ID:          "note-file-rb-1",
						Title:       "Updated",
						Version:     2,
						FilePath:    noteFile,
						ContentJSON: newContent,
						ContentHash: newHash,
					},
					{
						ID:          "note-file-rb-2",
						Title:       "Triggers Rollback",
						Version:     1,
						ContentHash: "h2",
					},
				},
			}
			_, err = uFail.Call(ctx, rollbackBatch)
			assert.ErrorIs(t, err, failErr)

			diskBytes, readErr := os.ReadFile(noteFile)
			require.NoError(t, readErr)
			assert.Equal(t, originalContent, string(diskBytes))
		},
	)

	t.Run(
		"preserves FilePath and CreatedAt on metadata-only note update",
		func(t *testing.T) {
			testDB, u := setupSyncTestEnv(t)
			ctx := context.Background()

			noteFile := filepath.Join(t.TempDir(), "meta-note.json")
			content := `{"ops":[{"insert":"keep me"}]}`
			sum := sha256.Sum256([]byte(content))
			hash := hex.EncodeToString(sum[:])

			createBatch := &model.SyncChangesBatch{
				SermonNotes: []*entity.SermonNote{
					{
						ID:          "note-meta-preserve",
						Title:       "Before",
						Version:     1,
						FilePath:    noteFile,
						ContentJSON: content,
						ContentHash: hash,
					},
				},
			}
			_, err := u.Call(ctx, createBatch)
			require.NoError(t, err)

			var created entity.SermonNote
			require.NoError(
				t,
				testDB.First(&created, "id = ?", "note-meta-preserve").Error,
			)

			updateBatch := &model.SyncChangesBatch{
				SermonNotes: []*entity.SermonNote{
					{
						ID:          "note-meta-preserve",
						Title:       "After",
						Version:     2,
						ContentHash: hash,
					},
				},
			}
			res, err := u.Call(ctx, updateBatch)
			require.NoError(t, err)
			assert.Empty(t, res.MissingContentIDs)

			var updated entity.SermonNote
			require.NoError(
				t,
				testDB.First(&updated, "id = ?", "note-meta-preserve").Error,
			)
			assert.Equal(t, "After", updated.Title)
			assert.Equal(t, noteFile, updated.FilePath)
			assert.False(t, updated.CreatedAt.IsZero())
		},
	)
}
