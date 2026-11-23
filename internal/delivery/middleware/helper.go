package middleware

import "github.com/labstack/echo/v5"

const (
	RequestIDKey = "request_id"
)

func SetRequestID(ctx *echo.Context, requestID string) {
	ctx.Set(RequestIDKey, requestID)
}

func GetRequestID(ctx *echo.Context) string {
	return ctx.Get(RequestIDKey).(string)
}
