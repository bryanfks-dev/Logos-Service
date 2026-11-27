package uploadsermonnoteattachmentfile_test

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
	sermonnote "github.com/bryanfks-dev/logos/internal/repository/sermon_note"
	uploadsermonnoteattachmentfile "github.com/bryanfks-dev/logos/internal/usecase/upload_sermon_note_attachment_file"
	"github.com/bryanfks-dev/logos/pkg/constant"
	"github.com/bryanfks-dev/logos/pkg/entity"
	entityerror "github.com/bryanfks-dev/logos/pkg/error/entity"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAttachmentUploadTestEnv(
	t *testing.T,
) (
	*gorm.DB,
	*uploadsermonnoteattachmentfile.UploadSermonNoteAttachmentFileUsecase,
	string,
) {
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
	sermonNoteRepo := sermonnote.NewSermonNoteRepository(
		testDB,
		fsManager,
		log,
	)

	appPath := t.TempDir()
	v := viper.New()
	v.Set("app.path", appPath)

	u := uploadsermonnoteattachmentfile.NewUploadSermonNoteAttachmentFileUsecase(
		&uploadsermonnoteattachmentfile.UploadSermonNoteAttachmentFileUsecaseOption{
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

			res, err = u.Call(
				ctx,
				&uploadsermonnoteattachmentfile.UploadSermonNoteAttachmentFileInput{
					AttachmentID:   uuid.Nil(),
					AttachmentHash: "abc",
					File:           bytes.NewReader([]byte("data")),
				},
			)
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

			missingID := uuid.NewV7()
			res, err := u.Call(
				ctx,
				&uploadsermonnoteattachmentfile.UploadSermonNoteAttachmentFileInput{
					AttachmentID:   missingID,
					AttachmentHash: "abc",
					File:           bytes.NewReader([]byte("data")),
				},
			)
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
				StoragePath:    "",
				ChecksumSHA256: "hash",
				DeletedAt:      gorm.DeletedAt{Time: time.Now(), Valid: true},
			}
			require.NoError(t, testDB.Create(deletedAtt).Error)

			res, err = u.Call(
				ctx,
				&uploadsermonnoteattachmentfile.UploadSermonNoteAttachmentFileInput{
					AttachmentID:   deletedID,
					AttachmentHash: "hash",
					File:           bytes.NewReader([]byte("data")),
				},
			)
			assert.Nil(t, res)
			assert.ErrorIs(
				t,
				err,
				entityerror.ErrSermonNoteAttachmentNotFound(deletedAtt.ID),
			)
		},
	)

	t.Run(
		"skips error log when attachment is not found, logs error on retrieval failure",
		func(t *testing.T) {
			core, recorded := observer.New(zap.ErrorLevel)
			obsLog := zap.New(core)

			dbName := fmt.Sprintf(
				"file:attachment_upload_obs_%s?mode=memory&cache=shared",
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
			v := viper.New()
			v.Set("app.path", t.TempDir())

			uc := uploadsermonnoteattachmentfile.NewUploadSermonNoteAttachmentFileUsecase(
				&uploadsermonnoteattachmentfile.UploadSermonNoteAttachmentFileUsecaseOption{
					SermonNoteRepository: repo,
					FsManager:            fsManager,
					Config:               v,
					Log:                  obsLog,
				},
			)

			missingID := uuid.NewV7()
			input := &uploadsermonnoteattachmentfile.UploadSermonNoteAttachmentFileInput{
				AttachmentID:   missingID,
				AttachmentHash: "abc",
				File:           bytes.NewReader([]byte("data")),
			}
			res, err := uc.Call(ctx, input)
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

			res, err = uc.Call(ctx, input)
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
		"returns 200 immediately when already uploaded with matching hash",
		func(t *testing.T) {
			testDB, u, appPath := setupAttachmentUploadTestEnv(t)

			note := &entity.SermonNote{
				ID:    uuid.NewV7().String(),
				Title: "Note",
			}
			require.NoError(t, testDB.Create(note).Error)

			attID := uuid.NewV7()
			att := &entity.SermonNoteAttachment{
				ID:       attID.String(),
				NoteID:   note.ID,
				FileName: "bulletin.pdf",
				MimeType: "application/pdf",
				FileSize: 10,
				StoragePath: filepath.Join(
					appPath,
					"2026",
					attID.String()+".pdf",
				),
				ChecksumSHA256: "deadbeef",
				Status:         constant.UploadStatusUploaded,
			}
			require.NoError(t, testDB.Create(att).Error)

			res, err := u.Call(
				ctx,
				&uploadsermonnoteattachmentfile.UploadSermonNoteAttachmentFileInput{
					AttachmentID:   attID,
					AttachmentHash: "DEADBEEF",
					File:           bytes.NewReader([]byte("ignored")),
				},
			)
			require.NoError(t, err)
			assert.Equal(t, att.ID, res.AttachmentID)
			assert.Equal(t, "deadbeef", res.AttachmentChecksumSHA256)
			assert.Equal(t, constant.UploadStatusStored, res.Status)

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

			note := &entity.SermonNote{
				ID:    uuid.NewV7().String(),
				Title: "Note",
			}
			require.NoError(t, testDB.Create(note).Error)

			attID := uuid.NewV7()
			att := &entity.SermonNoteAttachment{
				ID:             attID.String(),
				NoteID:         note.ID,
				FileName:       "audio.m4a",
				MimeType:       "audio/m4a",
				FileSize:       12,
				StoragePath:    "",
				ChecksumSHA256: "0000000000000000",
				Status:         constant.UploadStatusStored,
			}
			require.NoError(t, testDB.Create(att).Error)

			res, err := u.Call(
				ctx,
				&uploadsermonnoteattachmentfile.UploadSermonNoteAttachmentFileInput{
					AttachmentID:   attID,
					AttachmentHash: "wrong-hash",
					File:           bytes.NewReader([]byte("corrupt bytes")),
				},
			)
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
				constant.UploadStatusStored,
				failedAtt.Status,
			)
		},
	)

	t.Run(
		"streams, verifies hash, moves atomically, and updates DB",
		func(t *testing.T) {
			testDB, u, appPath := setupAttachmentUploadTestEnv(t)

			note := &entity.SermonNote{
				ID:    uuid.NewV7().String(),
				Title: "Note",
			}
			require.NoError(t, testDB.Create(note).Error)

			payload := []byte("PDF binary stream payload")
			sum := sha256.Sum256(payload)
			hashHex := hex.EncodeToString(sum[:])

			createdAt := time.Date(2026, 11, 25, 10, 0, 0, 0, time.UTC)
			attID := uuid.NewV7()
			att := &entity.SermonNoteAttachment{
				ID:             attID.String(),
				NoteID:         note.ID,
				FileName:       "Sunday_Bulletin.pdf",
				MimeType:       "application/pdf",
				FileSize:       int64(len(payload)),
				StoragePath:    "",
				ChecksumSHA256: hashHex,
				Status:         constant.UploadStatusStored,
				CreatedAt:      createdAt,
			}
			require.NoError(t, testDB.Create(att).Error)

			res, err := u.Call(
				ctx,
				&uploadsermonnoteattachmentfile.UploadSermonNoteAttachmentFileInput{
					AttachmentID:   attID,
					AttachmentHash: hashHex,
					File:           bytes.NewReader(payload),
				},
			)
			require.NoError(t, err)
			assert.Equal(t, att.ID, res.AttachmentID)
			assert.Equal(t, hashHex, res.AttachmentChecksumSHA256)
			assert.Equal(t, constant.UploadStatusUploaded, res.Status)

			expectedPath := filepath.Join(
				appPath,
				"notes",
				"2026",
				"11",
				att.ID+".pdf",
			)
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
