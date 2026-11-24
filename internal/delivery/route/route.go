package route

import (
	"net/http"

	"github.com/bryanfks-dev/selah/internal/delivery"
	"github.com/labstack/echo/v5"
	echoSwagger "github.com/swaggo/echo-swagger/v2"
)

type RouteConfig struct {
	App                  *echo.Echo
	EtcController        *delivery.EtcController
	ErrorController      *delivery.ErrorController
	SermonNoteController *delivery.SermonNoteController
}

func NewRouteConfig(
	app *echo.Echo,
	etcController *delivery.EtcController,
	errorController *delivery.ErrorController,
	sermonNoteController *delivery.SermonNoteController,
) *RouteConfig {
	return &RouteConfig{
		App:                  app,
		EtcController:        etcController,
		ErrorController:      errorController,
		SermonNoteController: sermonNoteController,
	}
}

func (c *RouteConfig) Setup() {
	c.SetupSwaggerRoute()
	c.SetupEtcRoute()
	c.SetupSermonNoteRoute()
	c.SetupErrorHandler()
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

func (c *RouteConfig) SetupSermonNoteRoute() {
	if c.SermonNoteController == nil {
		return
	}

	api := c.App.Group("/api")
	api.POST("/sync/notes", c.SermonNoteController.SyncNotes)
}

func (c *RouteConfig) SetupErrorHandler() {
	c.App.HTTPErrorHandler = c.ErrorController.HTTPErrorHandler
}
