package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bryanfks-dev/selah/internal/delivery/middleware"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestLogRequestMiddleware(t *testing.T) {
	e := echo.New()
	e.Use(middleware.LogRequestMiddleware())
	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "logged")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "logged", rec.Body.String())
}
