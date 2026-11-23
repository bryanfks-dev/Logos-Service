package route

import (
	"net/http"

	"github.com/bryanfks-dev/selah/internal/delivery"
	"github.com/labstack/echo/v5"
	echoSwagger "github.com/swaggo/echo-swagger/v2"
)

type RouteConfig struct {
	App           *echo.Echo
	EtcController *delivery.EtcController
}

func NewRouteConfig(app *echo.Echo, etcController *delivery.EtcController) *RouteConfig {
	return &RouteConfig{
		App:           app,
		EtcController: etcController,
	}
}

func (c *RouteConfig) Setup() {
	c.SetupSwaggerRoute()
	c.SetupEtcRoute()
}

func (c *RouteConfig) SetupSwaggerRoute() {
	c.App.GET("/swagger", func(ctx *echo.Context) error {
		return ctx.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	c.App.GET("/swagger/*", echoSwagger.WrapHandler)
}

func (c *RouteConfig) SetupEtcRoute() {
	api := c.App.Group("/api")
	api.GET("/healthz", c.EtcController.HealthCheck)
}
