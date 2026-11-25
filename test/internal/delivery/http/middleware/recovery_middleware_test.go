package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bryanfks-dev/logos/internal/delivery/http/middleware"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestRecoveryMiddleware(t *testing.T) {
	e := echo.New()
	e.Use(middleware.RecoveryMiddleware())
	e.GET("/panic", func(c *echo.Context) error {
		panic("simulated disaster")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		e.ServeHTTP(rec, req)
	})

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
