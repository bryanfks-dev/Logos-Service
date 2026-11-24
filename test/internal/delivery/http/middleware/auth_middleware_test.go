package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	deliveryhttp "github.com/bryanfks-dev/selah/internal/delivery/http"
	"github.com/bryanfks-dev/selah/internal/delivery/http/middleware"
	"github.com/bryanfks-dev/selah/internal/delivery/http/middleware/correlation"
	"github.com/bryanfks-dev/selah/internal/model"
	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAuthMiddleware(t *testing.T) {
	const validToken = "my-secret-token-12345"

	newConfiguredViper := func(token string) *viper.Viper {
		v := viper.New()
		v.Set("web.token", token)
		return v
	}

	setupServer := func(mw echo.MiddlewareFunc) *echo.Echo {
		e := echo.New()
		errorController := deliveryhttp.NewErrorController(zap.NewNop())
		e.HTTPErrorHandler = errorController.HTTPErrorHandler

		e.Use(middleware.RequestIDMiddleware())
		e.Use(mw)

		e.GET("/protected", func(ctx *echo.Context) error {
			return ctx.String(http.StatusOK, "authorized access")
		})

		return e
	}

	t.Run("allows request with valid X-Token", func(t *testing.T) {
		v := newConfiguredViper(validToken)
		e := setupServer(middleware.AuthMiddleware(v))

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("X-Token", validToken)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "authorized access", rec.Body.String())
	})

	t.Run("allows request with lowercase x-token header", func(t *testing.T) {
		v := newConfiguredViper(validToken)
		e := setupServer(middleware.AuthMiddleware(v))

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("x-token", validToken)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "authorized access", rec.Body.String())
	})

	t.Run(
		"allows request with whitespace trimmed around X-Token",
		func(t *testing.T) {
			v := newConfiguredViper(validToken)
			e := setupServer(middleware.AuthMiddleware(v))

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("X-Token", "   "+validToken+"   ")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "authorized access", rec.Body.String())
		},
	)

	t.Run(
		"rejects request with Authorization header because only X-Token is accepted",
		func(t *testing.T) {
			v := newConfiguredViper(validToken)
			e := setupServer(middleware.AuthMiddleware(v))

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", "Bearer "+validToken)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		},
	)

	t.Run(
		"rejects request with X-API-Key header because only X-Token is accepted",
		func(t *testing.T) {
			v := newConfiguredViper(validToken)
			e := setupServer(middleware.AuthMiddleware(v))

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("X-API-Key", validToken)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		},
	)

	t.Run(
		"rejects request with missing X-Token header with 401",
		func(t *testing.T) {
			v := newConfiguredViper(validToken)
			e := setupServer(middleware.AuthMiddleware(v))

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("X-Request-ID", "req-unauth-1")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)

			var res model.APIResponse[any]
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.False(t, res.Success)
			require.NotNil(t, res.Error)
			assert.Equal(t, "unauthorized", res.Error.Code)
			assert.Equal(t, "req-unauth-1", res.Meta.RequestID)
		},
	)

	t.Run(
		"rejects request with invalid X-Token with 401",
		func(t *testing.T) {
			v := newConfiguredViper(validToken)
			e := setupServer(middleware.AuthMiddleware(v))

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("X-Token", "invalid-wrong-token")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)

			var res model.APIResponse[any]
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.False(t, res.Success)
			assert.Equal(t, "unauthorized", res.Error.Code)
		},
	)

	t.Run(
		"allows request when server web.token is empty string",
		func(t *testing.T) {
			v := newConfiguredViper("")
			e := setupServer(middleware.AuthMiddleware(v))

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("X-Token", anyToken())
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
		},
	)

	t.Run("allows request when nil viper is passed", func(t *testing.T) {
		mw := middleware.AuthMiddleware(nil)
		assert.NotNil(t, mw)

		e := setupServer(mw)
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("X-Token", validToken)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run(
		"allows request when web.token key is not set in viper",
		func(t *testing.T) {
			emptyViper := viper.New()
			mw := middleware.AuthMiddleware(emptyViper)
			assert.NotNil(t, mw)

			e := setupServer(mw)
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("X-Token", validToken)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
		},
	)

	t.Run(
		"preserves correlation request ID on unauthorized response",
		func(t *testing.T) {
			v := newConfiguredViper(validToken)
			e := setupServer(middleware.AuthMiddleware(v))

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("X-Request-ID", "custom-trace-id-777")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.Equal(
				t,
				"custom-trace-id-777",
				rec.Header().Get("X-Request-ID"),
			)

			var res model.APIResponse[any]
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.Equal(t, "custom-trace-id-777", res.Meta.RequestID)
		},
	)

	t.Run(
		"context correlation is accessible when auth succeeds",
		func(t *testing.T) {
			v := newConfiguredViper(validToken)
			e := echo.New()
			var capturedRequestID string

			e.Use(middleware.RequestIDMiddleware())
			e.Use(middleware.AuthMiddleware(v))
			e.GET("/captured", func(c *echo.Context) error {
				capturedRequestID = correlation.RequestID(c)
				return c.String(http.StatusOK, "ok")
			})

			req := httptest.NewRequest(http.MethodGet, "/captured", nil)
			req.Header.Set("X-Token", validToken)
			req.Header.Set("X-Request-ID", "captured-req-id-888")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "captured-req-id-888", capturedRequestID)
		},
	)
}

func anyToken() string {
	return "any-token-string"
}
