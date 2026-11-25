package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"uuid"

	"github.com/bryanfks-dev/logos/db"
	deliveryhttp "github.com/bryanfks-dev/logos/internal/delivery/http"
	"github.com/bryanfks-dev/logos/internal/delivery/http/middleware"
	"github.com/bryanfks-dev/logos/internal/model"
	"github.com/bryanfks-dev/logos/internal/repository"
	"github.com/bryanfks-dev/logos/internal/usecase"
	"github.com/bryanfks-dev/logos/pkg/entity"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSyncHandlerTestEnv(
	t *testing.T,
) (*echo.Echo, *deliveryhttp.SyncHandler, *gorm.DB) {
	dbName := fmt.Sprintf(
		"file:sync_handler_test_%s?mode=memory&cache=shared",
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

	syncUsecase := usecase.NewSyncUsecase(&usecase.SyncUsecaseOption{
		TxManager:              txManager,
		SermonNoteRepository:   sermonNoteRepository,
		GatheringRepository:    gatheringRepository,
		SermonSeriesRepository: sermonSeriesRepository,
		PreacherRepository:     preacherRepository,
		TagRepository:          tagRepository,
		SyncOutboxRepository:   syncOutboxRepository,
		Log:                    log,
	})

	syncHandler := deliveryhttp.NewSyncHandler(syncUsecase, log)

	app := echo.New()
	errorController := deliveryhttp.NewErrorController(log)
	app.HTTPErrorHandler = errorController.HTTPErrorHandler
	app.Use(middleware.RequestIDMiddleware())

	v1 := app.Group("/api/v1")
	v1.POST("/sync/push", syncHandler.Push)
	v1.GET("/sync/pull", syncHandler.Pull)

	return app, syncHandler, testDB
}

func TestSyncHandler_Push(t *testing.T) {
	app, handler, _ := setupSyncHandlerTestEnv(t)

	t.Run(
		"returns 200 with standard api response envelope and generated request id",
		func(t *testing.T) {
			reqBody := model.SyncPushRequest{
				Notes: []model.SyncNoteItem{
					{
						ID:          "push-note-1",
						Title:       "Push Note Title",
						Version:     1,
						ContentHash: "hash-p-1",
					},
				},
				Gatherings: []model.SyncGatheringItem{
					{
						ID:       "push-g-1",
						Name:     "Sunday Service",
						ColorHex: "#333333",
					},
				},
			}

			bodyBytes, err := json.Marshal(reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/sync/push",
				bytes.NewReader(bodyBytes),
			)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)

			var res model.APIResponse[model.SyncPushResponse]
			err = json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)

			assert.True(t, res.Success)
			require.NotNil(t, res.Data)
			assert.False(t, res.Data.SyncedAt.IsZero())
			assert.Contains(t, res.Data.MissingContentIDs, "push-note-1")
			assert.Empty(t, res.Data.Conflicts)
			assert.NotEmpty(t, res.Meta.RequestID)
		},
	)

	t.Run("propagates custom request id from header", func(t *testing.T) {
		customReqID := "custom-trace-req-id-12345"
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/sync/push",
			bytes.NewReader([]byte("{}")),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Request-ID", customReqID)
		rec := httptest.NewRecorder()

		app.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var res model.APIResponse[model.SyncPushResponse]
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		require.NoError(t, err)

		assert.True(t, res.Success)
		assert.Equal(t, customReqID, res.Meta.RequestID)
	})

	t.Run("returns 400 for malformed json request body", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/sync/push",
			bytes.NewReader([]byte("{invalid-json")),
		)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		app.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var res model.APIResponse[any]
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		require.NoError(t, err)

		assert.False(t, res.Success)
		require.NotNil(t, res.Error)
		assert.Equal(t, "invalid_payload", res.Error.Code)
	})

	t.Run(
		"returns 400 with field errors when validation fails",
		func(t *testing.T) {
			reqBody := model.SyncPushRequest{
				Notes: []model.SyncNoteItem{
					{ID: "push-note-invalid", Title: ""},
				},
				Gatherings: []model.SyncGatheringItem{
					{ID: "push-g-invalid", Name: "Sunday", ColorHex: "bad"},
				},
			}

			bodyBytes, err := json.Marshal(reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/sync/push",
				bytes.NewReader(bodyBytes),
			)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var res model.APIResponse[any]
			err = json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)

			assert.False(t, res.Success)
			require.NotNil(t, res.Error)
			assert.Equal(t, "invalid_payload", res.Error.Code)
			assert.NotEmpty(t, res.Error.Errors)
		},
	)

	t.Run(
		"constructors instantiate handler and controller alias",
		func(t *testing.T) {
			assert.NotNil(t, handler)
			ctrl := deliveryhttp.NewSyncController(nil, zap.NewNop())
			assert.NotNil(t, ctrl)
		},
	)
}

func TestSyncHandler_Pull(t *testing.T) {
	app, _, testDB := setupSyncHandlerTestEnv(t)

	now := time.Now()
	g := &entity.Gathering{
		ID:        "g-pull-test-1",
		Name:      "Sunday Worship",
		UpdatedAt: now,
	}
	require.NoError(t, testDB.Create(g).Error)

	t.Run(
		"returns 200 with datetime server_time when since is omitted",
		func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/sync/pull", nil)
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)

			var res model.APIResponse[model.SyncPullResponse]
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)

			assert.True(t, res.Success)
			require.NotNil(t, res.Data)
			assert.False(t, res.Data.ServerTime.IsZero())

			// Verify JSON response has server_time in datetime format (RFC3339 string, not a number)
			var raw map[string]any
			err = json.Unmarshal(rec.Body.Bytes(), &raw)
			require.NoError(t, err)
			dataMap, ok := raw["data"].(map[string]any)
			require.True(t, ok)
			serverTimeStr, ok := dataMap["server_time"].(string)
			require.True(
				t,
				ok,
				"server_time should be serialized as string datetime, not unix int",
			)
			assert.NotEmpty(t, serverTimeStr)

			require.NotEmpty(t, res.Data.Gatherings)
			assert.Equal(t, "g-pull-test-1", res.Data.Gatherings[0].ID)
		},
	)

	t.Run(
		"returns 400 when since query param is unix integer",
		func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/sync/pull?since=0",
				nil,
			)
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var res model.APIResponse[any]
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.False(t, res.Success)
			require.NotNil(t, res.Error)
			assert.Equal(t, "invalid_payload", res.Error.Code)
		},
	)

	t.Run(
		"returns 200 with date-only since query param",
		func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/sync/pull?since=2026-01-01",
				nil,
			)
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)

			var res model.APIResponse[model.SyncPullResponse]
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.True(t, res.Success)
		},
	)

	t.Run(
		"returns 200 with RFC3339 datetime since query param",
		func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/sync/pull?since=2026-01-01T00:00:00Z",
				nil,
			)
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)

			var res model.APIResponse[model.SyncPullResponse]
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.True(t, res.Success)
		},
	)

	t.Run(
		"returns 400 when since query param is invalid format",
		func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/sync/pull?since=not-a-timestamp",
				nil,
			)
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var res model.APIResponse[any]
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.False(t, res.Success)
			require.NotNil(t, res.Error)
			assert.Equal(t, "invalid_payload", res.Error.Code)
		},
	)
}
