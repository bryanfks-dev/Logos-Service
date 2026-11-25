package middleware

import (
	"encoding/hex"
	"uuid"

	"github.com/bryanfks-dev/logos/internal/delivery/http/middleware/correlation"
	"github.com/labstack/echo/v5"
)

func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx *echo.Context) error {
			requestID := ctx.Request().Header.Get("X-Request-ID")
			if requestID == "" {
				randomUUID := uuid.New()
				requestID = hex.EncodeToString(randomUUID[:])
			}

			correlation.WithRequestID(ctx, requestID)
			ctx.Response().Header().Set("X-Request-ID", requestID)

			return next(ctx)
		}
	}
}
