package correlation

import "github.com/labstack/echo/v5"

const (
	RequestIDKey = "request_id"
)

func WithRequestID(ctx *echo.Context, requestID string) {
	ctx.Set(RequestIDKey, requestID)
}

func GetRequestID(ctx *echo.Context) string {
	if val := ctx.Get(RequestIDKey); val != nil {
		if id, ok := val.(string); ok {
			return id
		}
	}
	return ""
}
