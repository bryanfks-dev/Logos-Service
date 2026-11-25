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
			reqID := ctx.Request().Header.Get("X-Request-ID")
			if reqID == "" {
				randomUUID := uuid.New()
				reqID = hex.EncodeToString(randomUUID[:])
			}

			correlation.WithRequestID(ctx, reqID)
			ctx.Response().Header().Set("X-Request-ID", reqID)

			return next(ctx)
		}
	}
}
