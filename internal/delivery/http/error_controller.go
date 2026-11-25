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
	apperror.ErrKindNotFound:        http.StatusNotFound,
	apperror.ErrKindBadRequest:      http.StatusBadRequest,
	apperror.ErrKindConflict:        http.StatusConflict,
	apperror.ErrKindUnavailable:     http.StatusServiceUnavailable,
	apperror.ErrKindExternalService: http.StatusBadGateway,
	apperror.ErrKindInternal:        http.StatusInternalServerError,
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

	reqID := correlation.GetRequestID(ctx)
	stat := http.StatusInternalServerError
	code := "code_unavailable"

	var statusCoder echo.HTTPStatusCoder
	var appErr *apperror.AppError
	var validationErrs validation.Errors
	var fieldErrors []model.FieldError

	switch {
	case errors.Is(err, echo.ErrUnauthorized):
		stat = http.StatusUnauthorized
		code = "unauthorized"
	case errors.Is(err, echo.ErrNotFound):
		stat = http.StatusNotFound
		code = "not_found"
	case errors.Is(err, echo.ErrMethodNotAllowed):
		stat = http.StatusMethodNotAllowed
		code = "method_not_allowed"
	case errors.Is(err, echo.ErrUnsupportedMediaType):
		stat = http.StatusUnsupportedMediaType
		code = "invalid_media_type"
	case errors.As(err, &validationErrs):
		stat = http.StatusUnprocessableEntity
		code = "validation_error"
		fieldErrors = toFieldErrors(validationErrs)
	case errors.As(err, &appErr):
		if s, ok := errKindStatus[appErr.Kind]; ok {
			stat = s
		}
		code = appErr.Code
	case errors.As(err, &statusCoder):
		stat = statusCoder.StatusCode()
		code = "code_unavailable"
		if stat >= 500 {
			c.log.Error(
				"unexpected error",
				zap.Error(err),
				zap.String("request_id", reqID),
			)
		}
	default:
		stat = http.StatusInternalServerError
		code = "code_unavailable"
		c.log.Error(
			"unexpected error",
			zap.Error(err),
			zap.String("request_id", reqID),
		)
	}

	apiErr := model.APIError{
		Code:   code,
		Errors: fieldErrors,
	}

	res := builder.NewAPIResponseBuilder[any]().
		SetRequestID(reqID).
		SetError(apiErr).
		Build()

	_ = ctx.JSON(stat, res)
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
