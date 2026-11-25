package correlation_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bryanfks-dev/logos/internal/delivery/http/middleware/correlation"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestRequestID(t *testing.T) {
	e := echo.New()

	t.Run(
		"context without request id returns empty string",
		func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			assert.Equal(t, "", correlation.RequestID(ctx))
			assert.Equal(t, "", correlation.GetRequestID(ctx))
		},
	)

	t.Run("context with request id returns the id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		correlation.WithRequestID(ctx, "custom-req-id-123")

		assert.Equal(t, "custom-req-id-123", correlation.RequestID(ctx))
		assert.Equal(t, "custom-req-id-123", correlation.GetRequestID(ctx))
	})

	t.Run(
		"context with non-string request id returns empty string",
		func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			ctx.Set("request_id", 12345)

			assert.Equal(t, "", correlation.RequestID(ctx))
		},
	)
}
