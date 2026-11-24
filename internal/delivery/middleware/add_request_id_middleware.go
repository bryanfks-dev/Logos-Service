package middleware

import (
	"encoding/hex"
	"uuid"

	"github.com/bryanfks-dev/selah/internal/delivery/middleware/correlation"
	"github.com/labstack/echo/v5"
	echomiddleware "github.com/labstack/echo/v5/middleware"
)

func AddRequestIDMiddleware() echo.MiddlewareFunc {
	config := echomiddleware.RequestIDConfig{
		Generator: func() string {
			id := uuid.New()
			return hex.EncodeToString(id[:])
		},
		RequestIDHandler: func(ctx *echo.Context, requestID string) {
			correlation.WithRequestID(ctx, requestID)
		},
	}
	return echomiddleware.RequestIDWithConfig(config)
}
