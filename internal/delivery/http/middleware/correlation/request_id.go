package correlation

import "github.com/labstack/echo/v5"

const requestIDKey = "request_id"

func WithRequestID(ctx *echo.Context, requestID string) {
	ctx.Set(requestIDKey, requestID)
}

func RequestID(ctx *echo.Context) string {
	if val := ctx.Get(requestIDKey); val != nil {
		if id, ok := val.(string); ok {
			return id
		}
	}

	return ""
}

func GetRequestID(ctx *echo.Context) string {
	return RequestID(ctx)
}
