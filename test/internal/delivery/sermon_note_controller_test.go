package delivery_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bryanfks-dev/selah/internal/delivery"
	middlewarecorrelation "github.com/bryanfks-dev/selah/internal/delivery/middleware/correlation"
	"github.com/bryanfks-dev/selah/internal/model"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockSyncUsecase struct {
	syncFn func(ctx context.Context, req *model.SyncNotesRequest) (*model.SyncNotesResponse, error)
}

func (m *mockSyncUsecase) SyncNotes(
	ctx context.Context,
	req *model.SyncNotesRequest,
) (*model.SyncNotesResponse, error) {
	if m.syncFn != nil {
		return m.syncFn(ctx, req)
	}

	return &model.SyncNotesResponse{
		SyncedAt:   time.Now().UTC(),
		AppliedIDs: []string{"note-1"},
	}, nil
}

func TestSermonNoteController(t *testing.T) {
	log := zap.NewNop()

	t.Run("SyncNotes returns 200 on success", func(t *testing.T) {
		mockUC := &mockSyncUsecase{
			syncFn: func(ctx context.Context, req *model.SyncNotesRequest) (*model.SyncNotesResponse, error) {
				return &model.SyncNotesResponse{
					SyncedAt:   time.Now().UTC(),
					AppliedIDs: []string{"note-123"},
					Notes: []model.SyncNoteItem{
						{ID: "note-123", Title: "Test Note"},
					},
				}, nil
			},
		}

		controller := delivery.NewSermonNoteController(mockUC, log)

		e := echo.New()
		body := model.SyncNotesRequest{
			Notes: []model.SyncNoteItem{
				{ID: "note-123", Title: "Test Note"},
			},
		}
		bodyBytes, err := json.Marshal(body)
		require.NoError(t, err)

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/sync/notes",
			bytes.NewReader(bodyBytes),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		middlewarecorrelation.WithRequestID(ctx, "sync-req-123")

		err = controller.SyncNotes(ctx)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var res model.ApiResponse[model.SyncNotesResponse]
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		require.NoError(t, err)
		assert.True(t, res.Success)
		assert.Equal(t, "sync-req-123", res.Meta.RequestID)
		assert.NotNil(t, res.Data)
		assert.Equal(t, []string{"note-123"}, res.Data.AppliedIDs)
	})

	t.Run("SyncNotes returns 400 on invalid JSON body", func(t *testing.T) {
		controller := delivery.NewSermonNoteController(&mockSyncUsecase{}, log)

		e := echo.New()
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/sync/notes",
			bytes.NewBufferString("{invalid json"),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		middlewarecorrelation.WithRequestID(ctx, "sync-req-400")

		err := controller.SyncNotes(ctx)
		assert.Error(t, err)

		var httpErr *echo.HTTPError
		require.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	})

	t.Run("SyncNotes propagates usecase error", func(t *testing.T) {
		expectedErr := errors.New("database sync failure")
		mockUC := &mockSyncUsecase{
			syncFn: func(ctx context.Context, req *model.SyncNotesRequest) (*model.SyncNotesResponse, error) {
				return nil, expectedErr
			},
		}

		controller := delivery.NewSermonNoteController(mockUC, log)

		e := echo.New()
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/sync/notes",
			bytes.NewBufferString(`{"notes":[]}`),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		middlewarecorrelation.WithRequestID(ctx, "sync-req-500")

		err := controller.SyncNotes(ctx)
		assert.ErrorIs(t, err, expectedErr)
	})
}
