package http_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
	"uuid"

	"github.com/bryanfks-dev/logos/infra/fs"
	deliveryhttp "github.com/bryanfks-dev/logos/internal/delivery/http"
	"github.com/bryanfks-dev/logos/internal/delivery/http/middleware"
	"github.com/bryanfks-dev/logos/internal/model"
	"github.com/bryanfks-dev/logos/internal/model/response"
	sermonnote "github.com/bryanfks-dev/logos/internal/repository/sermon_note"
	uploadsermonnoteattachmentfile "github.com/bryanfks-dev/logos/internal/usecase/upload_sermon_note_attachment_file"
	"github.com/bryanfks-dev/logos/pkg/constant"
	"github.com/bryanfks-dev/logos/pkg/entity"
	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSermonNoteControllerTestEnv(
	t *testing.T,
) (*echo.Echo, *gorm.DB, string) {
	dbName := fmt.Sprintf(
		"file:sermon_note_ctrl_test_%s?mode=memory&cache=shared",
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

	uploadUsecase := uploadsermonnoteattachmentfile.NewUploadSermonNoteAttachmentFileUsecase(
		&uploadsermonnoteattachmentfile.UploadSermonNoteAttachmentFileUsecaseOption{
			SermonNoteRepository: sermonNoteRepo,
			FsManager:            fsManager,
			Config:               v,
			Log:                  log,
		},
	)

	controller := deliveryhttp.NewSermonNoteController(uploadUsecase, log)

	app := echo.New()
	errorController := deliveryhttp.NewErrorController(log)
	app.HTTPErrorHandler = errorController.HTTPErrorHandler
	app.Use(middleware.RequestIDMiddleware())

	api := app.Group("/api")
	api.PUT("/attachments/:id/file", controller.UploadAttachmentFile)

	return app, testDB, appPath
}

func TestSermonNoteController_UploadAttachmentFile(t *testing.T) {
	t.Run("returns 404 when attachment is not found", func(t *testing.T) {
		app, _, _ := setupSermonNoteControllerTestEnv(t)

		req := httptest.NewRequest(
			http.MethodPut,
			"/api/attachments/missing-id/file",
			bytes.NewReader([]byte("bytes")),
		)
		req.Header.Set("X-Content-Hash", "abc")
		rec := httptest.NewRecorder()

		app.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("returns 400 invalid_hash when hash mismatches", func(t *testing.T) {
		app, testDB, _ := setupSermonNoteControllerTestEnv(t)

		note := &entity.SermonNote{ID: "ctrl-note-1", Title: "Note"}
		require.NoError(t, testDB.Create(note).Error)

		att := &entity.SermonNoteAttachment{
			ID:             "ctrl-att-bad",
			NoteID:         note.ID,
			FileName:       "doc.pdf",
			MimeType:       "application/pdf",
			FileSize:       5,
			ChecksumSHA256: "expected-hash",
		}
		require.NoError(t, testDB.Create(att).Error)

		req := httptest.NewRequest(
			http.MethodPut,
			"/api/attachments/"+att.ID+"/file",
			bytes.NewReader([]byte("wrong")),
		)
		req.Header.Set("X-Content-Hash", "expected-hash")
		rec := httptest.NewRecorder()

		app.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var res model.APIResponse[any]
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
		require.NotNil(t, res.Error)
		assert.Equal(t, "invalid_hash", res.Error.Code)
	})

	t.Run(
		"returns 200 and stores file atomically on valid upload",
		func(t *testing.T) {
			app, testDB, appPath := setupSermonNoteControllerTestEnv(t)

			note := &entity.SermonNote{ID: "ctrl-note-2", Title: "Note"}
			require.NoError(t, testDB.Create(note).Error)

			payload := []byte("sermon audio binary content")
			sum := sha256.Sum256(payload)
			hashHex := hex.EncodeToString(sum[:])

			att := &entity.SermonNoteAttachment{
				ID:             "ctrl-att-ok",
				NoteID:         note.ID,
				FileName:       "Sunday_Bulletin.pdf",
				MimeType:       "application/pdf",
				FileSize:       int64(len(payload)),
				ChecksumSHA256: hashHex,
				CreatedAt:      time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC),
			}
			require.NoError(t, testDB.Create(att).Error)

			req := httptest.NewRequest(
				http.MethodPut,
				"/api/attachments/"+att.ID+"/file",
				bytes.NewReader(payload),
			)
			req.Header.Set("Content-Type", "application/octet-stream")
			req.Header.Set("X-Content-Hash", hashHex)
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)

			var res model.APIResponse[response.UploadSermonNoteAttachmentFileResponse]
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
			require.NotNil(t, res.Data)
			assert.Equal(t, att.ID, res.Data.ID)
			assert.Equal(t, hashHex, res.Data.ChecksumSHA256)
			assert.Equal(t, constant.UploadStatusUploaded, res.Data.Status)

			stored, err := os.ReadFile(
				filepath.Join(appPath, "notes", "2026", "11", att.ID+".pdf"),
			)
			require.NoError(t, err)
			assert.Equal(t, payload, stored)
		},
	)
}
