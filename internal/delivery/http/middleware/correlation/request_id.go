package correlation

import "github.com/labstack/echo/v5"

const requestIDKey = "request_id"

func WithRequestID(ctx *echo.Context, reqID string) {
	ctx.Set(requestIDKey, reqID)
}

func GetRequestID(ctx *echo.Context) string {
	if val := ctx.Get(requestIDKey); val != nil {
		if id, ok := val.(string); ok {
			return id
		}
	}

	return ""
}
