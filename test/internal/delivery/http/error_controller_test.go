package http_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	deliveryhttp "github.com/bryanfks-dev/logos/internal/delivery/http"
	"github.com/bryanfks-dev/logos/internal/delivery/http/middleware/correlation"
	"github.com/bryanfks-dev/logos/internal/model"
	apperror "github.com/bryanfks-dev/logos/pkg/error"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestErrorController(t *testing.T) {
	log := zap.NewNop()
	controller := deliveryhttp.NewErrorController(log)

	setupContext := func(
		reqID string,
	) (*echo.Echo, *echo.Context, *httptest.ResponseRecorder) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		correlation.WithRequestID(ctx, reqID)
		return e, ctx, rec
	}

	t.Run(
		"when response already committed does not write response",
		func(t *testing.T) {
			_, ctx, rec := setupContext("req-1")
			res, _ := echo.UnwrapResponse(ctx.Response())
			res.Committed = true

			controller.HTTPErrorHandler(ctx, echo.ErrNotFound)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Empty(t, rec.Body.String())
		},
	)

	t.Run("handles echo.ErrUnauthorized", func(t *testing.T) {
		_, ctx, rec := setupContext("req-unauthorized")
		controller.HTTPErrorHandler(ctx, echo.ErrUnauthorized)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var apiRes model.APIResponse[any]
		err := json.Unmarshal(rec.Body.Bytes(), &apiRes)
		require.NoError(t, err)
		assert.False(t, apiRes.Success)
		assert.Equal(t, "unauthorized", apiRes.Error.Code)
		assert.Equal(t, "req-unauthorized", apiRes.Meta.RequestID)
	})

	t.Run("handles echo.ErrNotFound", func(t *testing.T) {
		_, ctx, rec := setupContext("req-not-found")
		controller.HTTPErrorHandler(ctx, echo.ErrNotFound)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		var apiRes model.APIResponse[any]
		err := json.Unmarshal(rec.Body.Bytes(), &apiRes)
		require.NoError(t, err)
		assert.False(t, apiRes.Success)
		assert.Equal(t, "not_found", apiRes.Error.Code)
		assert.Equal(t, "req-not-found", apiRes.Meta.RequestID)
	})

	t.Run("handles echo.ErrMethodNotAllowed", func(t *testing.T) {
		_, ctx, rec := setupContext("req-method-na")
		controller.HTTPErrorHandler(ctx, echo.ErrMethodNotAllowed)

		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
		var apiRes model.APIResponse[any]
		err := json.Unmarshal(rec.Body.Bytes(), &apiRes)
		require.NoError(t, err)
		assert.False(t, apiRes.Success)
		assert.Equal(t, "method_not_allowed", apiRes.Error.Code)
	})

	t.Run("handles echo.ErrUnsupportedMediaType", func(t *testing.T) {
		_, ctx, rec := setupContext("req-media-type")
		controller.HTTPErrorHandler(ctx, echo.ErrUnsupportedMediaType)

		assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)
		var apiRes model.APIResponse[any]
		err := json.Unmarshal(rec.Body.Bytes(), &apiRes)
		require.NoError(t, err)
		assert.False(t, apiRes.Success)
		assert.Equal(t, "invalid_media_type", apiRes.Error.Code)
	})

	t.Run("handles validation.Errors with field details", func(t *testing.T) {
		_, ctx, rec := setupContext("req-validation")
		verr := validation.Errors{
			"notes": validation.Errors{
				"0": validation.Errors{
					"title": errors.New("cannot be blank"),
				},
			},
		}

		controller.HTTPErrorHandler(ctx, verr)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var apiRes model.APIResponse[any]
		err := json.Unmarshal(rec.Body.Bytes(), &apiRes)
		require.NoError(t, err)
		assert.False(t, apiRes.Success)
		require.NotNil(t, apiRes.Error)
		assert.Equal(t, "invalid_payload", apiRes.Error.Code)
		require.Len(t, apiRes.Error.Errors, 1)
		assert.Equal(t, "notes.0.title", apiRes.Error.Errors[0].Field)
		assert.Equal(t, "cannot be blank", apiRes.Error.Errors[0].Message)
	})

	t.Run("handles AppError kinds", func(t *testing.T) {
		tests := []struct {
			appErr     *apperror.AppError
			wantStatus int
			wantCode   string
		}{
			{
				apperror.NewErrNotFound("resource_not_found", "not found"),
				http.StatusNotFound,
				"resource_not_found",
			},
			{
				apperror.NewErrConflict("resource_conflict", "conflict"),
				http.StatusConflict,
				"resource_conflict",
			},
			{
				apperror.NewErrMethodNotAllowed("not_allowed", "not allowed"),
				http.StatusMethodNotAllowed,
				"not_allowed",
			},
			{
				apperror.NewErrUnavailable("service_down", "down", nil),
				http.StatusServiceUnavailable,
				"service_down",
			},
			{
				apperror.NewErrExternalService("gateway_err", "gateway", nil),
				http.StatusBadGateway,
				"gateway_err",
			},
			{
				apperror.NewErrInternal("db_error", "db error", nil),
				http.StatusInternalServerError,
				"db_error",
			},
			{
				&apperror.AppError{
					Kind: apperror.ErrKind(999),
					Code: "unknown_kind",
				},
				http.StatusInternalServerError,
				"unknown_kind",
			},
		}

		for _, tt := range tests {
			_, ctx, rec := setupContext("req-app-err")
			controller.HTTPErrorHandler(ctx, tt.appErr)

			assert.Equal(t, tt.wantStatus, rec.Code)
			var apiRes model.APIResponse[any]
			err := json.Unmarshal(rec.Body.Bytes(), &apiRes)
			require.NoError(t, err)
			assert.False(t, apiRes.Success)
			assert.Equal(t, tt.wantCode, apiRes.Error.Code)
		}
	})

	t.Run("handles echo.HTTPStatusCoder (4xx and 5xx)", func(t *testing.T) {
		_, ctx400, rec400 := setupContext("req-400")
		controller.HTTPErrorHandler(
			ctx400,
			echo.NewHTTPError(http.StatusBadRequest, "bad request"),
		)
		assert.Equal(t, http.StatusBadRequest, rec400.Code)

		_, ctx500, rec500 := setupContext("req-500")
		controller.HTTPErrorHandler(
			ctx500,
			echo.NewHTTPError(
				http.StatusInternalServerError,
				"something crashed",
			),
		)
		assert.Equal(t, http.StatusInternalServerError, rec500.Code)
	})

	t.Run("handles generic unhandled error", func(t *testing.T) {
		_, ctx, rec := setupContext("req-generic")
		controller.HTTPErrorHandler(
			ctx,
			errors.New("unexpected catastrophic error"),
		)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		var apiRes model.APIResponse[any]
		err := json.Unmarshal(rec.Body.Bytes(), &apiRes)
		require.NoError(t, err)
		assert.False(t, apiRes.Success)
		assert.Equal(t, "code_unavailable", apiRes.Error.Code)
	})
}
