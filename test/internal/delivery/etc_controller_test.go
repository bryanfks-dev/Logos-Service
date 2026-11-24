package delivery_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bryanfks-dev/selah/internal/delivery"
	middlewarecorrelation "github.com/bryanfks-dev/selah/internal/delivery/middleware/correlation"
	"github.com/bryanfks-dev/selah/internal/model"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEtcController(t *testing.T) {
	controller := delivery.NewEtcController()

	t.Run("HealthCheck returns 200 and request id", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		middlewarecorrelation.WithRequestID(ctx, "health-check-req-123")

		err := controller.HealthCheck(ctx)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var res model.ApiResponse[any]
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		require.NoError(t, err)
		assert.True(t, res.Success)
		assert.Equal(t, "health-check-req-123", res.Meta.RequestID)
		assert.False(t, res.Meta.Timestamp.IsZero())
	})
}
