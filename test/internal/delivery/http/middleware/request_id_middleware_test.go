package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bryanfks-dev/logos/internal/delivery/http/middleware"
	"github.com/bryanfks-dev/logos/internal/delivery/http/middleware/correlation"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestRequestIDMiddleware(t *testing.T) {
	t.Run("generates new request id if none provided", func(t *testing.T) {
		e := echo.New()
		var capturedRequestID string

		e.Use(middleware.RequestIDMiddleware())
		e.GET("/test", func(ctx *echo.Context) error {
			capturedRequestID = correlation.RequestID(ctx)
			return ctx.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEmpty(t, capturedRequestID)
		assert.Equal(t, capturedRequestID, rec.Header().Get("X-Request-ID"))
	})

	t.Run("reuses existing X-Request-ID header", func(t *testing.T) {
		e := echo.New()
		var capturedRequestID string

		e.Use(middleware.RequestIDMiddleware())
		e.GET("/test", func(ctx *echo.Context) error {
			capturedRequestID = correlation.RequestID(ctx)
			return ctx.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-ID", "custom-client-id-456")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "custom-client-id-456", capturedRequestID)
		assert.Equal(
			t,
			"custom-client-id-456",
			rec.Header().Get("X-Request-ID"),
		)
	})
}
