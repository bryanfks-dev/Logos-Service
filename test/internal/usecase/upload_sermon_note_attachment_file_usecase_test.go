package usecase_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
	"uuid"

	"github.com/bryanfks-dev/logos/infra/fs"
	"github.com/bryanfks-dev/logos/internal/model"
	"github.com/bryanfks-dev/logos/internal/repository"
	"github.com/bryanfks-dev/logos/internal/usecase"
	"github.com/bryanfks-dev/logos/pkg/constant"
	"github.com/bryanfks-dev/logos/pkg/entity"
	entityerror "github.com/bryanfks-dev/logos/pkg/error/entity"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAttachmentUploadTestEnv(
	t *testing.T,
) (*gorm.DB, *usecase.UploadSermonNoteAttachmentFileUsecase, string) {
	dbName := fmt.Sprintf(
		"file:attachment_upload_test_%s?mode=memory&cache=shared",
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
	sermonNoteRepo := repository.NewSermonNoteRepository(
		testDB,
		fsManager,
		log,
	)

	appPath := t.TempDir()
	v := viper.New()
	v.Set("app.path", appPath)

	u := usecase.NewUploadSermonNoteAttachmentFileUsecase(
		&usecase.UploadSermonNoteAttachmentFileUsecaseOption{
			SermonNoteRepository: sermonNoteRepo,
			FsManager:            fsManager,
			Config:               v,
			Log:                  log,
		},
	)

	return testDB, u, appPath
}

func TestUploadSermonNoteAttachmentFileUsecase_Call(t *testing.T) {
	ctx := context.Background()

	t.Run(
		"returns ErrSermonNoteAttachmentInvalidFile when input is invalid",
		func(t *testing.T) {
			_, u, _ := setupAttachmentUploadTestEnv(t)

			res, err := u.Call(ctx, nil)
			assert.Nil(t, res)
			assert.ErrorIs(
				t,
				err,
				entityerror.ErrSermonNoteAttachmentInvalidFile(),
			)
		},
	)

	t.Run(
		"returns 404 when attachment does not exist or is soft deleted",
		func(t *testing.T) {
			testDB, u, _ := setupAttachmentUploadTestEnv(t)

			res, err := u.Call(ctx, &model.UploadSermonNoteAttachmentFileInput{
				ID:          "missing-att",
				ContentHash: "abc",
				File:        bytes.NewReader([]byte("data")),
			})
			assert.Nil(t, res)
			assert.ErrorIs(
				t,
				err,
				entityerror.ErrSermonNoteAttachmentNotFound("missing-att"),
			)

			note := &entity.SermonNote{ID: "note-del-att", Title: "Note"}
			require.NoError(t, testDB.Create(note).Error)

			deletedAtt := &entity.SermonNoteAttachment{
				ID:             "att-deleted",
				NoteID:         note.ID,
				FileName:       "bulletin.pdf",
				MimeType:       "application/pdf",
				FileSize:       4,
				StoragePath:    "",
				ChecksumSHA256: "hash",
				DeletedAt:      gorm.DeletedAt{Time: time.Now(), Valid: true},
			}
			require.NoError(t, testDB.Create(deletedAtt).Error)

			res, err = u.Call(ctx, &model.UploadSermonNoteAttachmentFileInput{
				ID:          deletedAtt.ID,
				ContentHash: "hash",
				File:        bytes.NewReader([]byte("data")),
			})
			assert.Nil(t, res)
			assert.ErrorIs(
				t,
				err,
				entityerror.ErrSermonNoteAttachmentNotFound(deletedAtt.ID),
			)
		},
	)

	t.Run(
		"returns 200 immediately when already uploaded with matching hash",
		func(t *testing.T) {
			testDB, u, appPath := setupAttachmentUploadTestEnv(t)

			note := &entity.SermonNote{ID: "note-idem", Title: "Note"}
			require.NoError(t, testDB.Create(note).Error)

			att := &entity.SermonNoteAttachment{
				ID:             "att-idem",
				NoteID:         note.ID,
				FileName:       "bulletin.pdf",
				MimeType:       "application/pdf",
				FileSize:       10,
				StoragePath:    filepath.Join(appPath, "2026", "att-idem.pdf"),
				ChecksumSHA256: "deadbeef",
				Status:         constant.UploadStatusUploaded,
			}
			require.NoError(t, testDB.Create(att).Error)

			res, err := u.Call(ctx, &model.UploadSermonNoteAttachmentFileInput{
				ID:          att.ID,
				ContentHash: "DEADBEEF",
				File:        bytes.NewReader([]byte("ignored")),
			})
			require.NoError(t, err)
			assert.Equal(t, att.ID, res.ID)
			assert.Equal(t, "deadbeef", res.ChecksumSHA256)
			assert.Equal(t, "stored", res.Status)

			_, statErr := os.Stat(
				filepath.Join(appPath, ".tmp", att.ID+".part"),
			)
			assert.True(t, os.IsNotExist(statErr))
		},
	)

	t.Run(
		"deletes temp file and returns ErrInvalidHash on mismatch",
		func(t *testing.T) {
			testDB, u, appPath := setupAttachmentUploadTestEnv(t)

			note := &entity.SermonNote{ID: "note-badhash", Title: "Note"}
			require.NoError(t, testDB.Create(note).Error)

			att := &entity.SermonNoteAttachment{
				ID:             "att-badhash",
				NoteID:         note.ID,
				FileName:       "audio.m4a",
				MimeType:       "audio/m4a",
				FileSize:       12,
				StoragePath:    "",
				ChecksumSHA256: "0000000000000000",
				Status:         constant.UploadStatusNotUploaded,
			}
			require.NoError(t, testDB.Create(att).Error)

			res, err := u.Call(ctx, &model.UploadSermonNoteAttachmentFileInput{
				ID:          att.ID,
				ContentHash: "wrong-hash",
				File:        bytes.NewReader([]byte("corrupt bytes")),
			})
			assert.Nil(t, res)
			assert.ErrorIs(
				t,
				err,
				entityerror.ErrSermonNoteAttachmentInvalidHash(),
			)

			_, statErr := os.Stat(
				filepath.Join(appPath, ".tmp", att.ID+".part"),
			)
			assert.True(t, os.IsNotExist(statErr))

			var failedAtt entity.SermonNoteAttachment
			require.NoError(t, testDB.First(&failedAtt, "id = ?", att.ID).Error)
			assert.Equal(
				t,
				constant.UploadStatusNotUploaded,
				failedAtt.Status,
			)
		},
	)

	t.Run(
		"streams, verifies hash, moves atomically, and updates DB",
		func(t *testing.T) {
			testDB, u, appPath := setupAttachmentUploadTestEnv(t)

			note := &entity.SermonNote{ID: "note-ok", Title: "Note"}
			require.NoError(t, testDB.Create(note).Error)

			payload := []byte("PDF binary stream payload")
			sum := sha256.Sum256(payload)
			hashHex := hex.EncodeToString(sum[:])

			createdAt := time.Date(2026, 11, 25, 10, 0, 0, 0, time.UTC)
			att := &entity.SermonNoteAttachment{
				ID:             "att-ok",
				NoteID:         note.ID,
				FileName:       "Sunday_Bulletin.pdf",
				MimeType:       "application/pdf",
				FileSize:       int64(len(payload)),
				StoragePath:    "",
				ChecksumSHA256: hashHex,
				Status:         constant.UploadStatusNotUploaded,
				CreatedAt:      createdAt,
			}
			require.NoError(t, testDB.Create(att).Error)

			res, err := u.Call(ctx, &model.UploadSermonNoteAttachmentFileInput{
				ID:          att.ID,
				ContentHash: hashHex,
				File:        bytes.NewReader(payload),
			})
			require.NoError(t, err)
			assert.Equal(t, att.ID, res.ID)
			assert.Equal(t, hashHex, res.ChecksumSHA256)
			assert.Equal(t, "stored", res.Status)

			expectedPath := filepath.Join(appPath, "2026", att.ID+".pdf")
			storedBytes, err := os.ReadFile(expectedPath)
			require.NoError(t, err)
			assert.Equal(t, payload, storedBytes)

			_, tmpStatErr := os.Stat(
				filepath.Join(appPath, ".tmp", att.ID+".part"),
			)
			assert.True(t, os.IsNotExist(tmpStatErr))

			var updated entity.SermonNoteAttachment
			require.NoError(t, testDB.First(&updated, "id = ?", att.ID).Error)
			assert.Equal(
				t,
				constant.UploadStatusUploaded,
				updated.Status,
			)
			assert.Equal(t, expectedPath, updated.StoragePath)
		},
	)
}
