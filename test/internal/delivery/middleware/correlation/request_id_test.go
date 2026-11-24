package middlewarecorrelation_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	middlewarecorrelation "github.com/bryanfks-dev/selah/internal/delivery/middleware/correlation"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestRequestID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	t.Run(
		"get request id when not set returns empty string",
		func(t *testing.T) {
			assert.Empty(t, middlewarecorrelation.GetRequestID(ctx))
			assert.Empty(t, middlewarecorrelation.RequestID(ctx))
			assert.Empty(t, middlewarecorrelation.RequestID(nil))
		},
	)

	t.Run("set and get request id", func(t *testing.T) {
		middlewarecorrelation.WithRequestID(ctx, "test-request-id-123")
		assert.Equal(
			t,
			"test-request-id-123",
			middlewarecorrelation.GetRequestID(ctx),
		)
		assert.Equal(
			t,
			"test-request-id-123",
			middlewarecorrelation.RequestID(ctx),
		)
	})

	t.Run("get request id when value is not a string", func(t *testing.T) {
		ctx.Set("request_id", 12345)
		assert.Empty(t, middlewarecorrelation.GetRequestID(ctx))
		assert.Empty(t, middlewarecorrelation.RequestID(ctx))
	})
}
