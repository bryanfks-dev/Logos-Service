package middleware

import (
	"net/http"

	"github.com/labstack/echo/v5"
	echomiddleware "github.com/labstack/echo/v5/middleware"
)

func CorsMiddleware() echo.MiddlewareFunc {
	allowedMethods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodOptions,
	}
	allowedHeaders := []string{
		echo.HeaderOrigin,
		echo.HeaderContentType,
		echo.HeaderAccept,
		echo.HeaderAuthorization,
		echo.HeaderXRequestID,
	}
	unsafeAllowOriginFunc := func(ctx *echo.Context, origin string) (string, bool, error) {
		return origin, true, nil
	}

	config := echomiddleware.CORSConfig{
		AllowCredentials:      true,
		UnsafeAllowOriginFunc: unsafeAllowOriginFunc,
		AllowMethods:          allowedMethods,
		AllowHeaders:          allowedHeaders,
	}

	return echomiddleware.CORSWithConfig(config)
}
