package middleware

import (
	"crypto/subtle"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
)

const headerXToken = "X-Token"

func AuthMiddleware(config *viper.Viper) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx *echo.Context) error {
			if config == nil {
				return nil
			}

			expectedToken := strings.TrimSpace(config.GetString("web.token"))
			if expectedToken == "" {
				return nil
			}

			clientToken := strings.TrimSpace(
				ctx.Request().Header.Get(headerXToken),
			)
			if clientToken == "" {
				return echo.ErrUnauthorized
			}

			if subtle.ConstantTimeCompare(
				[]byte(clientToken),
				[]byte(expectedToken),
			) != 1 {
				return echo.ErrUnauthorized
			}

			return next(ctx)
		}
	}
}
