package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bryanfks-dev/selah/internal/delivery/middleware"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestCorsMiddleware(t *testing.T) {
	e := echo.New()
	e.Use(middleware.CorsMiddleware())
	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	t.Run("handles preflight OPTIONS request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set(echo.HeaderOrigin, "https://example.com")
		req.Header.Set(
			echo.HeaderAccessControlRequestMethod,
			http.MethodPost,
		)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Equal(
			t,
			"https://example.com",
			rec.Header().Get(echo.HeaderAccessControlAllowOrigin),
		)
		assert.Equal(
			t,
			"true",
			rec.Header().Get(echo.HeaderAccessControlAllowCredentials),
		)
		assert.Contains(
			t,
			rec.Header().Get(echo.HeaderAccessControlAllowMethods),
			http.MethodPost,
		)
		assert.Contains(
			t,
			rec.Header().Get(echo.HeaderAccessControlAllowHeaders),
			echo.HeaderContentType,
		)
	})

	t.Run("handles regular GET request with CORS header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(echo.HeaderOrigin, "https://example.com")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(
			t,
			"https://example.com",
			rec.Header().Get(echo.HeaderAccessControlAllowOrigin),
		)
		assert.Equal(
			t,
			"true",
			rec.Header().Get(echo.HeaderAccessControlAllowCredentials),
		)
	})
}
