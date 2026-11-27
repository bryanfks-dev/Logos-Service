package downloadsermonnoteattachmentfile_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
	"uuid"

	"github.com/bryanfks-dev/logos/infra/fs"
	sermonnote "github.com/bryanfks-dev/logos/internal/repository/sermon_note"
	downloadsermonnoteattachmentfile "github.com/bryanfks-dev/logos/internal/usecase/download_sermon_note_attachment_file"
	"github.com/bryanfks-dev/logos/pkg/constant"
	"github.com/bryanfks-dev/logos/pkg/entity"
	entityerror "github.com/bryanfks-dev/logos/pkg/error/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAttachmentDownloadTestEnv(
	t *testing.T,
) (
	*gorm.DB,
	*downloadsermonnoteattachmentfile.DownloadSermonNoteAttachmentFileUsecase,
	string,
) {
	dbName := fmt.Sprintf(
		"file:attachment_download_test_%s?mode=memory&cache=shared",
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

	log := zap.NewNop()
	fsManager := fs.NewManager()
	sermonNoteRepo := sermonnote.NewSermonNoteRepository(
		testDB,
		fsManager,
		log,
	)

	u := downloadsermonnoteattachmentfile.NewDownloadSermonNoteAttachmentFileUsecase(
		sermonNoteRepo,
		fsManager,
		log,
	)

	return testDB, u, t.TempDir()
}

func TestDownloadSermonNoteAttachmentFileUsecase_Call(t *testing.T) {
	ctx := context.Background()

	t.Run(
		"returns ErrSermonNoteAttachmentNotFound when attachment id is nil",
		func(t *testing.T) {
			_, u, _ := setupAttachmentDownloadTestEnv(t)

			res, err := u.Call(ctx, uuid.Nil())
			assert.Nil(t, res)
			assert.ErrorIs(
				t,
				err,
				entityerror.ErrSermonNoteAttachmentNotFound(
					uuid.Nil().String(),
				),
			)
		},
	)

	t.Run(
		"returns 404 when attachment does not exist or is soft deleted",
		func(t *testing.T) {
			testDB, u, _ := setupAttachmentDownloadTestEnv(t)

			missingID := uuid.NewV7()
			res, err := u.Call(ctx, missingID)
			assert.Nil(t, res)
			assert.ErrorIs(
				t,
				err,
				entityerror.ErrSermonNoteAttachmentNotFound(missingID.String()),
			)

			note := &entity.SermonNote{
				ID:    uuid.NewV7().String(),
				Title: "Note",
			}
			require.NoError(t, testDB.Create(note).Error)

			deletedID := uuid.NewV7()
			deletedAtt := &entity.SermonNoteAttachment{
				ID:             deletedID.String(),
				NoteID:         note.ID,
				FileName:       "bulletin.pdf",
				MimeType:       "application/pdf",
				FileSize:       4,
				StoragePath:    "/tmp/missing.pdf",
				ChecksumSHA256: "hash",
				Status:         constant.UploadStatusUploaded,
				DeletedAt:      gorm.DeletedAt{Time: time.Now(), Valid: true},
			}
			require.NoError(t, testDB.Create(deletedAtt).Error)

			res, err = u.Call(ctx, deletedID)
			assert.Nil(t, res)
			assert.ErrorIs(
				t,
				err,
				entityerror.ErrSermonNoteAttachmentNotFound(deletedAtt.ID),
			)
		},
	)

	t.Run(
		"does not log error when attachment is not found, logs error on retrieval failure",
		func(t *testing.T) {
			core, recorded := observer.New(zap.ErrorLevel)
			obsLog := zap.New(core)

			dbName := fmt.Sprintf(
				"file:attachment_download_obs_%s?mode=memory&cache=shared",
				uuid.NewV7().String(),
			)
			obsDB, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
			require.NoError(t, err)
			require.NoError(
				t,
				obsDB.AutoMigrate(&entity.SermonNoteAttachment{}),
			)

			fsManager := fs.NewManager()
			repo := sermonnote.NewSermonNoteRepository(obsDB, fsManager, obsLog)
			uc := downloadsermonnoteattachmentfile.NewDownloadSermonNoteAttachmentFileUsecase(
				repo,
				fsManager,
				obsLog,
			)

			missingID := uuid.NewV7()
			res, err := uc.Call(ctx, missingID)
			assert.Nil(t, res)
			assert.ErrorIs(
				t,
				err,
				entityerror.ErrSermonNoteAttachmentNotFound(missingID.String()),
			)
			assert.Empty(t, recorded.All())

			sqlDB, err := obsDB.DB()
			require.NoError(t, err)
			require.NoError(t, sqlDB.Close())

			res, err = uc.Call(ctx, uuid.NewV7())
			assert.Nil(t, res)
			assert.ErrorIs(
				t,
				err,
				entityerror.ErrSermonNoteAttachmentRetrievalFailed(nil),
			)
			assert.Len(t, recorded.All(), 1)
		},
	)

	t.Run(
		"returns 404 when attachment is not uploaded yet or storage path is empty",
		func(t *testing.T) {
			testDB, u, _ := setupAttachmentDownloadTestEnv(t)

			note := &entity.SermonNote{
				ID:    uuid.NewV7().String(),
				Title: "Note",
			}
			require.NoError(t, testDB.Create(note).Error)

			storedID := uuid.NewV7()
			storedAtt := &entity.SermonNoteAttachment{
				ID:             storedID.String(),
				NoteID:         note.ID,
				FileName:       "bulletin.pdf",
				MimeType:       "application/pdf",
				FileSize:       10,
				StoragePath:    "",
				ChecksumSHA256: "abc",
				Status:         constant.UploadStatusStored,
			}
			require.NoError(t, testDB.Create(storedAtt).Error)

			res, err := u.Call(ctx, storedID)
			assert.Nil(t, res)
			assert.ErrorIs(
				t,
				err,
				entityerror.ErrSermonNoteAttachmentNotFound(storedAtt.ID),
			)

			emptyPathID := uuid.NewV7()
			uploadedEmptyPath := &entity.SermonNoteAttachment{
				ID:             emptyPathID.String(),
				NoteID:         note.ID,
				FileName:       "bulletin.pdf",
				MimeType:       "application/pdf",
				FileSize:       10,
				StoragePath:    "",
				ChecksumSHA256: "abc",
				Status:         constant.UploadStatusUploaded,
			}
			require.NoError(t, testDB.Create(uploadedEmptyPath).Error)

			res, err = u.Call(ctx, emptyPathID)
			assert.Nil(t, res)
			assert.ErrorIs(
				t,
				err,
				entityerror.ErrSermonNoteAttachmentNotFound(
					uploadedEmptyPath.ID,
				),
			)
		},
	)

	t.Run(
		"returns 404 when file is missing from storage disk",
		func(t *testing.T) {
			testDB, u, tmpDir := setupAttachmentDownloadTestEnv(t)

			note := &entity.SermonNote{
				ID:    uuid.NewV7().String(),
				Title: "Note",
			}
			require.NoError(t, testDB.Create(note).Error)

			attID := uuid.NewV7()
			att := &entity.SermonNoteAttachment{
				ID:             attID.String(),
				NoteID:         note.ID,
				FileName:       "missing.pdf",
				MimeType:       "application/pdf",
				FileSize:       10,
				StoragePath:    filepath.Join(tmpDir, "nonexistent.pdf"),
				ChecksumSHA256: "abc",
				Status:         constant.UploadStatusUploaded,
			}
			require.NoError(t, testDB.Create(att).Error)

			res, err := u.Call(ctx, attID)
			assert.Nil(t, res)
			assert.ErrorIs(
				t,
				err,
				entityerror.ErrSermonNoteAttachmentNotFound(att.ID),
			)
		},
	)

	t.Run(
		"returns ErrSermonNoteAttachmentRetrievalFailed when storage path is invalid",
		func(t *testing.T) {
			testDB, u, tmpDir := setupAttachmentDownloadTestEnv(t)

			note := &entity.SermonNote{
				ID:    uuid.NewV7().String(),
				Title: "Note",
			}
			require.NoError(t, testDB.Create(note).Error)

			attID := uuid.NewV7()
			att := &entity.SermonNoteAttachment{
				ID:             attID.String(),
				NoteID:         note.ID,
				FileName:       "dir.pdf",
				MimeType:       "application/pdf",
				FileSize:       10,
				StoragePath:    tmpDir,
				ChecksumSHA256: "abc",
				Status:         constant.UploadStatusUploaded,
			}
			require.NoError(t, testDB.Create(att).Error)

			res, err := u.Call(ctx, attID)
			assert.Nil(t, res)
			assert.ErrorIs(
				t,
				err,
				entityerror.ErrSermonNoteAttachmentRetrievalFailed(nil),
			)
		},
	)

	t.Run(
		"returns open file stream and metadata for uploaded attachment",
		func(t *testing.T) {
			testDB, u, tmpDir := setupAttachmentDownloadTestEnv(t)

			note := &entity.SermonNote{
				ID:    uuid.NewV7().String(),
				Title: "Note",
			}
			require.NoError(t, testDB.Create(note).Error)

			payload := []byte("sermon audio content bytes")
			attID := uuid.NewV7()
			filePath := filepath.Join(
				tmpDir,
				"notes",
				"2026",
				"11",
				attID.String()+".m4a",
			)
			require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0o755))
			require.NoError(t, os.WriteFile(filePath, payload, 0o644))

			updatedAt := time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC)
			att := &entity.SermonNoteAttachment{
				ID:             attID.String(),
				NoteID:         note.ID,
				FileName:       "Sunday_Sermon.m4a",
				MimeType:       "audio/mp4",
				FileSize:       0,
				StoragePath:    filePath,
				ChecksumSHA256: "deadbeef",
				Status:         constant.UploadStatusUploaded,
				UpdatedAt:      updatedAt,
			}
			require.NoError(t, testDB.Create(att).Error)

			res, err := u.Call(ctx, attID)
			require.NoError(t, err)
			require.NotNil(t, res)
			defer res.File.Close()

			assert.Equal(t, "Sunday_Sermon.m4a", res.FileName)
			assert.Equal(t, "audio/mp4", res.MimeType)
			assert.Equal(t, int64(len(payload)), res.FileSize)
			assert.Equal(t, "deadbeef", res.ChecksumSHA256)
			assert.True(t, res.UpdatedAt.Equal(updatedAt))

			gotBytes, err := io.ReadAll(res.File)
			require.NoError(t, err)
			assert.Equal(t, payload, gotBytes)
		},
	)
}
