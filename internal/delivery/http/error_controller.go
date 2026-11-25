package http

import (
	"errors"
	"net/http"

	"github.com/bryanfks-dev/logos/internal/delivery/http/middleware/correlation"
	"github.com/bryanfks-dev/logos/internal/model"
	"github.com/bryanfks-dev/logos/internal/model/builder"
	apperror "github.com/bryanfks-dev/logos/pkg/error"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

var errKindStatus = map[apperror.ErrKind]int{
	apperror.ErrKindNotFound:         http.StatusNotFound,
	apperror.ErrKindConflict:         http.StatusConflict,
	apperror.ErrKindMethodNotAllowed: http.StatusMethodNotAllowed,
	apperror.ErrKindUnavailable:      http.StatusServiceUnavailable,
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

	requestID := correlation.RequestID(ctx)
	status := http.StatusInternalServerError
	code := "code_unavailable"

	var statusCoder echo.HTTPStatusCoder
	var appErr *apperror.AppError
	var validationErrs validation.Errors
	var fieldErrors []model.FieldError

	switch {
	case errors.Is(err, echo.ErrUnauthorized):
		status = http.StatusUnauthorized
		code = "unauthorized"
	case errors.Is(err, echo.ErrNotFound):
		status = http.StatusNotFound
		code = "not_found"
	case errors.Is(err, echo.ErrMethodNotAllowed):
		status = http.StatusMethodNotAllowed
		code = "method_not_allowed"
	case errors.Is(err, echo.ErrUnsupportedMediaType):
		status = http.StatusUnsupportedMediaType
		code = "invalid_media_type"
	case errors.Is(err, apperror.ErrInvalidPayload):
		status = http.StatusBadRequest
		code = "invalid_payload"
	case errors.As(err, &validationErrs):
		status = http.StatusUnprocessableEntity
		code = "validation_error"
		fieldErrors = toFieldErrors(validationErrs)
	case errors.Is(err, apperror.ErrConflict):
		status = http.StatusConflict
		code = "conflict"
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

	apiErr := model.APIError{
		Code:   code,
		Errors: fieldErrors,
	}

	res := builder.NewAPIResponseBuilder[any]().
		SetRequestID(requestID).
		SetError(apiErr).
		Build()

	_ = ctx.JSON(status, res)
}

func toFieldErrors(errs validation.Errors) []model.FieldError {
	flat := errs.Flatten()
	fieldErrors := make([]model.FieldError, 0, len(flat))

	for _, fe := range flat {
		fieldErr := model.FieldError{
			Field:   fe.Field(),
			Message: fe.Message,
		}
		fieldErrors = append(fieldErrors, fieldErr)
	}

	return fieldErrors
}
