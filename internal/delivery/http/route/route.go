package route

import (
	nethttp "net/http"

	deliveryhttp "github.com/bryanfks-dev/selah/internal/delivery/http"
	"github.com/labstack/echo/v5"
	echoSwagger "github.com/swaggo/echo-swagger/v2"
)

type RouteConfig struct {
	App             *echo.Echo
	EtcController   *deliveryhttp.EtcController
	ErrorController *deliveryhttp.ErrorController
	SyncHandler     *deliveryhttp.SyncHandler
}

func NewRouteConfig(
	app *echo.Echo,
	etcController *deliveryhttp.EtcController,
	errorController *deliveryhttp.ErrorController,
	syncHandler *deliveryhttp.SyncHandler,
) *RouteConfig {
	return &RouteConfig{
		App:             app,
		EtcController:   etcController,
		ErrorController: errorController,
		SyncHandler:     syncHandler,
	}
}

func (c *RouteConfig) Setup() {
	c.SetupSwaggerRoute()
	c.SetupEtcRoute()
	c.SetupSyncRoute()
	c.SetupErrorHandler()
}

func (c *RouteConfig) SetupSwaggerRoute() {
	c.App.GET("/swagger", func(ctx *echo.Context) error {
		return ctx.Redirect(
			nethttp.StatusMovedPermanently,
			"/swagger/index.html",
		)
	})
	c.App.GET("/swagger/*", echoSwagger.WrapHandler)
}

func (c *RouteConfig) SetupEtcRoute() {
	api := c.App.Group("/api")
	api.GET("/healthz", c.EtcController.HealthCheck)
}

func (c *RouteConfig) SetupSyncRoute() {
	if c.SyncHandler != nil {
		v1 := c.App.Group("/api/v1")
		v1.POST("/sync/push", c.SyncHandler.Push)
		v1.GET("/sync/pull", c.SyncHandler.Pull)
	}
}

func (c *RouteConfig) SetupErrorHandler() {
	c.App.HTTPErrorHandler = c.ErrorController.HTTPErrorHandler
}
