package delivery

import (
	"errors"
	"net/http"

	middlewarecorrelation "github.com/bryanfks-dev/selah/internal/delivery/middleware/correlation"
	"github.com/bryanfks-dev/selah/internal/model"
	modelbuilder "github.com/bryanfks-dev/selah/internal/model/builder"
	apperror "github.com/bryanfks-dev/selah/pkg/error"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

var errKindStatus = map[apperror.ErrKind]int{
	apperror.ErrKindNotFound:         http.StatusNotFound,
	apperror.ErrKindConflict:         http.StatusConflict,
	apperror.ErrKindMethodNotAllowed: http.StatusMethodNotAllowed,
	apperror.ErrKindUnvailable:       http.StatusServiceUnavailable,
	apperror.ErrKindExternalService:  http.StatusBadGateway,
	apperror.ErrKindInternal:         http.StatusInternalServerError,
}

type ErrorController struct {
	log *zap.Logger
}

func NewErrorController(log *zap.Logger) *ErrorController {
	return &ErrorController{log: log}
}

func (c *ErrorController) HTTPErrorHandler(ctx *echo.Context, err error) {
	if r, _ := echo.UnwrapResponse(ctx.Response()); r != nil && r.Committed {
		return
	}

	requestID := middlewarecorrelation.GetRequestID(ctx)
	status := http.StatusInternalServerError
	code := "code_unavailable"

	var statusCoder echo.HTTPStatusCoder
	var appErr *apperror.AppError

	switch {
	case errors.Is(err, echo.ErrNotFound):
		status = http.StatusNotFound
		code = "not_found"
	case errors.Is(err, echo.ErrMethodNotAllowed):
		status = http.StatusMethodNotAllowed
		code = "method_not_allowed"
	case errors.Is(err, echo.ErrUnsupportedMediaType):
		status = http.StatusUnsupportedMediaType
		code = "invalid_media_type"
	case errors.As(err, &appErr):
		if stat, ok := errKindStatus[appErr.Kind]; ok {
			status = stat
		}
		code = appErr.Code
	case errors.As(err, &statusCoder):
		status = statusCoder.StatusCode()
		code = "code_unavailable"
		if status >= 500 {
			c.log.Error(
				"unexpected error",
				zap.Error(err),
				zap.String("request_id", requestID),
			)
		}
	default:
		status = http.StatusInternalServerError
		code = "code_unavailable"
		c.log.Error(
			"unexpected error",
			zap.Error(err),
			zap.String("request_id", requestID),
		)
	}

	res := modelbuilder.NewApiResponseBuilder[any]().
		SetRequestId(requestID).
		SetError(model.ApiError{Code: code}).
		Build()

	_ = ctx.JSON(status, res)
}
