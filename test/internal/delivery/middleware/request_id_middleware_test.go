package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bryanfks-dev/selah/internal/delivery/middleware"
	middlewarecorrelation "github.com/bryanfks-dev/selah/internal/delivery/middleware/correlation"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestRequestIDMiddleware(t *testing.T) {
	t.Run(
		"generates hex request id and populates context and response header",
		func(t *testing.T) {
			e := echo.New()
			e.Use(middleware.RequestIDMiddleware())

			var capturedID string
			e.GET("/test", func(c *echo.Context) error {
				capturedID = middlewarecorrelation.GetRequestID(c)
				return c.String(http.StatusOK, "ok")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.NotEmpty(t, capturedID)
			assert.Len(t, capturedID, 32)
			assert.Equal(
				t,
				capturedID,
				rec.Header().Get(echo.HeaderXRequestID),
			)
		},
	)

	t.Run("uses incoming request id when provided", func(t *testing.T) {
		e := echo.New()
		e.Use(middleware.RequestIDMiddleware())

		var capturedID string
		e.GET("/test", func(c *echo.Context) error {
			capturedID = middlewarecorrelation.GetRequestID(c)
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(echo.HeaderXRequestID, "custom-request-id-12345")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "custom-request-id-12345", capturedID)
		assert.Equal(
			t,
			"custom-request-id-12345",
			rec.Header().Get(echo.HeaderXRequestID),
		)
	})
}
