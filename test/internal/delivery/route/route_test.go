package route_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bryanfks-dev/selah/internal/delivery"
	middlewarecorrelation "github.com/bryanfks-dev/selah/internal/delivery/middleware/correlation"
	"github.com/bryanfks-dev/selah/internal/delivery/route"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRouteConfig(t *testing.T) {
	app := echo.New()
	etcController := delivery.NewEtcController()
	errorController := delivery.NewErrorController(zap.NewNop())
	sermonNoteController := delivery.NewSermonNoteController(nil, zap.NewNop())

	routeConfig := route.NewRouteConfig(
		app,
		etcController,
		errorController,
		sermonNoteController,
	)
	routeConfig.Setup()

	t.Run("GET /swagger redirects to /swagger/index.html", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/swagger", nil)
		rec := httptest.NewRecorder()

		app.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusMovedPermanently, rec.Code)
		assert.Equal(t, "/swagger/index.html", rec.Header().Get("Location"))
	})

	t.Run("GET /api/healthz returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
		rec := httptest.NewRecorder()

		// Simulate request id middleware injecting into context
		app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				middlewarecorrelation.WithRequestID(c, "route-test-req-id")
				return next(c)
			}
		})

		app.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run(
		"SetupErrorHandler configures app.HTTPErrorHandler",
		func(t *testing.T) {
			assert.NotNil(t, app.HTTPErrorHandler)
		},
	)
}
