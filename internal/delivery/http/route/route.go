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

type RouteConfigOption struct {
	App             *echo.Echo
	EtcController   *deliveryhttp.EtcController
	ErrorController *deliveryhttp.ErrorController
	SyncHandler     *deliveryhttp.SyncHandler
}

func NewRouteConfig(opt *RouteConfigOption) *RouteConfig {
	if opt == nil {
		opt = &RouteConfigOption{}
	}

	return &RouteConfig{
		App:             opt.App,
		EtcController:   opt.EtcController,
		ErrorController: opt.ErrorController,
		SyncHandler:     opt.SyncHandler,
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
	api := c.App.Group("/api")
	api.POST("/sync/push", c.SyncHandler.Push)
	api.GET("/sync/pull", c.SyncHandler.Pull)
}

func (c *RouteConfig) SetupErrorHandler() {
	c.App.HTTPErrorHandler = c.ErrorController.HTTPErrorHandler
}
