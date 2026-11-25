package route

import (
	nethttp "net/http"

	deliveryhttp "github.com/bryanfks-dev/logos/internal/delivery/http"
	"github.com/labstack/echo/v5"
	echoSwagger "github.com/swaggo/echo-swagger/v2"
)

type RouteConfig struct {
	App             *echo.Echo
	EtcController   *deliveryhttp.EtcController
	ErrorController *deliveryhttp.ErrorController
	SyncController  *deliveryhttp.SyncController
}

type RouteConfigOption struct {
	App             *echo.Echo
	EtcController   *deliveryhttp.EtcController
	ErrorController *deliveryhttp.ErrorController
	SyncController  *deliveryhttp.SyncController
}

func NewRouteConfig(opt *RouteConfigOption) *RouteConfig {
	if opt == nil {
		opt = &RouteConfigOption{}
	}

	return &RouteConfig{
		App:             opt.App,
		EtcController:   opt.EtcController,
		ErrorController: opt.ErrorController,
		SyncController:  opt.SyncController,
	}
}

func (c *RouteConfig) Setup() {
	c.setupSwaggerRoute()
	c.setupEtcRoute()
	c.setupSyncRoute()
	c.setupErrorHandler()
}

func (c *RouteConfig) setupSwaggerRoute() {
	c.App.GET("/swagger", func(ctx *echo.Context) error {
		return ctx.Redirect(
			nethttp.StatusMovedPermanently,
			"/swagger/index.html",
		)
	})
	c.App.GET("/swagger/*", echoSwagger.WrapHandler)
}

func (c *RouteConfig) setupEtcRoute() {
	api := c.App.Group("/api")
	api.GET("/healthz", c.EtcController.HealthCheck)
}

func (c *RouteConfig) setupSyncRoute() {
	api := c.App.Group("/api")
	api.POST("/sync/changes", c.SyncController.Push)
	api.GET("/sync/changes/:since", c.SyncController.Pull)
}

func (c *RouteConfig) setupErrorHandler() {
	c.App.HTTPErrorHandler = c.ErrorController.HTTPErrorHandler
}
