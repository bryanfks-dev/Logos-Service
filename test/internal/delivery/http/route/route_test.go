package route_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	deliveryhttp "github.com/bryanfks-dev/logos/internal/delivery/http"
	"github.com/bryanfks-dev/logos/internal/delivery/http/middleware/correlation"
	"github.com/bryanfks-dev/logos/internal/delivery/http/route"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRouteConfig(t *testing.T) {
	app := echo.New()
	etcController := deliveryhttp.NewEtcController()
	errorController := deliveryhttp.NewErrorController(zap.NewNop())

	routeConfig := route.NewRouteConfig(&route.RouteConfigOption{
		App:             app,
		EtcController:   etcController,
		ErrorController: errorController,
		SyncHandler:     nil,
	})
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

		app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				correlation.WithRequestID(c, "route-test-req-id")
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
