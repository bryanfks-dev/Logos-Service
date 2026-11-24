package middleware

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func GzipMiddleware() echo.MiddlewareFunc {
	return middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	})
}
