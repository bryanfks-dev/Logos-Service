package middleware

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func GzipMiddleware() echo.MiddlewareFunc {
	cfg := middleware.GzipConfig{
		Level: 5,
	}

	return middleware.GzipWithConfig(cfg)
}
