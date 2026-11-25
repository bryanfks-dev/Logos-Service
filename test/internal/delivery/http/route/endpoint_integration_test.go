package route_test

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bryanfks-dev/logos/di"
	_ "github.com/bryanfks-dev/logos/docs"
	deliveryhttp "github.com/bryanfks-dev/logos/internal/delivery/http"
	"github.com/bryanfks-dev/logos/internal/delivery/http/route"
	"github.com/bryanfks-dev/logos/internal/model"
	"github.com/bryanfks-dev/logos/internal/model/response"
	"github.com/bryanfks-dev/logos/pkg/config"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupIntegrationServer(t *testing.T) *echo.Echo {
	t.Helper()

	viperConfig := config.NewViper()
	logger := zap.NewNop()
	server := config.NewServer(viperConfig)

	etcController := deliveryhttp.NewEtcController()
	errorController := deliveryhttp.NewErrorController(logger)

	routeOpt := &route.RouteConfigOption{
		App:             server,
		EtcController:   etcController,
		ErrorController: errorController,
		SyncController:  nil,
	}
	routeConfig := route.NewRouteConfig(routeOpt)
	routeConfig.Setup()

	return server
}

func setupDIContainerServer(t *testing.T) *echo.Echo {
	t.Helper()

	viperConfig := config.NewViper()
	host := viperConfig.GetString("database.host")
	port := viperConfig.GetInt("database.port")
	conn, err := net.DialTimeout(
		"tcp",
		fmt.Sprintf("%s:%d", host, port),
		500*time.Millisecond,
	)
	if err != nil {
		t.Skipf("skipping DI container integration test with live db: %v", err)
	}
	_ = conn.Close()

	container, err := di.NewContainer()
	require.NoError(t, err)
	require.NotNil(t, container)

	container.RouteConfig.Setup()
	return container.App
}

func TestHealthCheckEndpoint_Integration(t *testing.T) {
	app := setupIntegrationServer(t)

	t.Run(
		"returns 200 with standard api response envelope and generated request id",
		func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(
				t,
				rec.Header().Get("Content-Type"),
				"application/json",
			)

			responseRequestID := rec.Header().Get(echo.HeaderXRequestID)
			assert.NotEmpty(t, responseRequestID)

			var res model.APIResponse[any]
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)

			assert.True(t, res.Success)
			assert.Nil(t, res.Data)
			assert.Nil(t, res.Error)
			assert.Equal(t, responseRequestID, res.Meta.RequestID)
			assert.False(t, res.Meta.Timestamp.IsZero())
			assert.WithinDuration(
				t,
				time.Now().UTC(),
				res.Meta.Timestamp,
				5*time.Second,
			)
		},
	)

	t.Run(
		"propagates custom request id from incoming header",
		func(t *testing.T) {
			customRequestID := "client-trace-req-998877"
			req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
			req.Header.Set(echo.HeaderXRequestID, customRequestID)
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(
				t,
				customRequestID,
				rec.Header().Get(echo.HeaderXRequestID),
			)

			var res model.APIResponse[any]
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)

			assert.True(t, res.Success)
			assert.Equal(t, customRequestID, res.Meta.RequestID)
		},
	)

	t.Run(
		"supports CORS preflight OPTIONS request on healthz",
		func(t *testing.T) {
			req := httptest.NewRequest(http.MethodOptions, "/api/healthz", nil)
			req.Header.Set(echo.HeaderOrigin, "https://app.logos.local")
			req.Header.Set(
				echo.HeaderAccessControlRequestMethod,
				http.MethodGet,
			)
			req.Header.Set(
				echo.HeaderAccessControlRequestHeaders,
				"Content-Type, X-Request-ID",
			)
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Equal(
				t,
				"https://app.logos.local",
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
				http.MethodGet,
			)
		},
	)

	t.Run(
		"includes CORS headers on regular GET request with Origin",
		func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
			req.Header.Set(echo.HeaderOrigin, "https://app.logos.local")
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(
				t,
				"https://app.logos.local",
				rec.Header().Get(echo.HeaderAccessControlAllowOrigin),
			)
			assert.Equal(
				t,
				"true",
				rec.Header().Get(echo.HeaderAccessControlAllowCredentials),
			)
		},
	)
}

func TestSwaggerEndpoints_Integration(t *testing.T) {
	app := setupIntegrationServer(t)

	t.Run("GET /swagger redirects to /swagger/index.html", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/swagger", nil)
		rec := httptest.NewRecorder()

		app.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusMovedPermanently, rec.Code)
		assert.Equal(t, "/swagger/index.html", rec.Header().Get("Location"))
	})

	t.Run(
		"GET /swagger/doc.json returns 200 with OpenAPI JSON",
		func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(
				t,
				rec.Header().Get("Content-Type"),
				"application/json",
			)

			var swaggerDoc map[string]any
			err := json.Unmarshal(rec.Body.Bytes(), &swaggerDoc)
			require.NoError(t, err)

			assert.Equal(t, "2.0", swaggerDoc["swagger"])
			info, ok := swaggerDoc["info"].(map[string]any)
			require.True(t, ok)
			assert.Equal(t, "Logos Service API", info["title"])
		},
	)
}

func TestNotFoundHandling_Integration(t *testing.T) {
	app := setupIntegrationServer(t)

	testCases := []struct {
		name string
		path string
	}{
		{
			name: "unknown api route",
			path: "/api/unknown-endpoint",
		},
		{
			name: "unknown nested route",
			path: "/api/v1/resource/123",
		},
		{
			name: "root non-existent path",
			path: "/non-existent",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code)
			assert.Contains(
				t,
				rec.Header().Get("Content-Type"),
				"application/json",
			)

			responseRequestID := rec.Header().Get(echo.HeaderXRequestID)
			assert.NotEmpty(t, responseRequestID)

			var res model.APIResponse[any]
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)

			assert.False(t, res.Success)
			assert.Nil(t, res.Data)
			require.NotNil(t, res.Error)
			assert.Equal(t, "not_found", res.Error.Code)
			assert.Equal(t, responseRequestID, res.Meta.RequestID)
		})
	}
}

func TestPanicRecovery_Integration(t *testing.T) {
	app := setupIntegrationServer(t)

	app.GET("/api/panic-test", func(c *echo.Context) error {
		panic("critical unexpected database connection panic")
	})

	t.Run(
		"recovers from panic and returns 500 structured error response",
		func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/panic-test", nil)
			rec := httptest.NewRecorder()

			assert.NotPanics(t, func() {
				app.ServeHTTP(rec, req)
			})

			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.Contains(
				t,
				rec.Header().Get("Content-Type"),
				"application/json",
			)

			responseRequestID := rec.Header().Get(echo.HeaderXRequestID)
			assert.NotEmpty(t, responseRequestID)

			var res model.APIResponse[any]
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)

			assert.False(t, res.Success)
			require.NotNil(t, res.Error)
			assert.Equal(t, "code_unavailable", res.Error.Code)
			assert.Equal(t, responseRequestID, res.Meta.RequestID)
		},
	)

	t.Run(
		"server continues to serve subsequent requests normally after panic",
		func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)

			var res model.APIResponse[any]
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.True(t, res.Success)
		},
	)
}

func TestDIContainerEndpoints_Integration(t *testing.T) {
	app := setupDIContainerServer(t)

	t.Run(
		"GET /api/healthz works end-to-end through full DI container",
		func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)

			var res model.APIResponse[any]
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)

			assert.True(t, res.Success)
			assert.NotEmpty(t, res.Meta.RequestID)
		},
	)

	t.Run(
		"GET /swagger redirects through full DI container",
		func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/swagger", nil)
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusMovedPermanently, rec.Code)
			assert.Equal(t, "/swagger/index.html", rec.Header().Get("Location"))
		},
	)

	t.Run(
		"GET /non-existent returns 404 through full DI container",
		func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/non-existent", nil)
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code)

			var res model.APIResponse[any]
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)

			assert.False(t, res.Success)
			assert.Equal(t, "not_found", res.Error.Code)
		},
	)

	t.Run(
		"POST /api/sync/changes returns 200 through full DI container",
		func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodPost,
				"/api/sync/changes",
				strings.NewReader("{}"),
			)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)

			var res model.APIResponse[response.SyncPushChangesResponse]
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)

			assert.True(t, res.Success)
			assert.NotEmpty(t, res.Meta.RequestID)
			require.NotNil(t, res.Data)
			assert.False(t, res.Data.SyncedAt.IsZero())
		},
	)

	t.Run(
		"GET /api/sync/changes/:since returns 200 through full DI container",
		func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodGet,
				"/api/sync/changes/2026-01-01T00:00:00Z",
				nil,
			)
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)

			var res model.APIResponse[response.SyncPullChangesResponse]
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			require.NoError(t, err)

			assert.True(t, res.Success)
			assert.NotEmpty(t, res.Meta.RequestID)
			require.NotNil(t, res.Data)
			assert.False(t, res.Data.ServerTime.IsZero())
		},
	)
}
