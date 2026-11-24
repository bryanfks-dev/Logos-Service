package middleware

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func LogRequestMiddleware() echo.MiddlewareFunc {
	return middleware.RequestLogger()
}
