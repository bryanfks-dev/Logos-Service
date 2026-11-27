package apperror_test

import (
	"errors"
	"testing"

	apperror "github.com/bryanfks-dev/logos/pkg/error"
	"github.com/stretchr/testify/assert"
)

func TestErrKind_IsDomain(t *testing.T) {
	assert.True(t, apperror.ErrKindNotFound.IsDomain())
	assert.True(t, apperror.ErrKindBadRequest.IsDomain())
	assert.True(t, apperror.ErrKindConflict.IsDomain())
	assert.False(t, apperror.ErrKindUnavailable.IsDomain())
	assert.False(t, apperror.ErrKindExternalService.IsDomain())
	assert.False(t, apperror.ErrKindInternal.IsDomain())
}

func TestAppError(t *testing.T) {
	cause := errors.New("root cause error")

	t.Run("Error method without cause", func(t *testing.T) {
		err := apperror.NewErrNotFound("not_found_code", "resource not found")
		assert.Equal(t, "not_found_code: resource not found", err.Error())
	})

	t.Run("Error method with cause", func(t *testing.T) {
		err := apperror.NewErrInternal(
			"db_err",
			"database failure",
			cause,
		)
		assert.Equal(
			t,
			"db_err: database failure: root cause error",
			err.Error(),
		)
	})

	t.Run("Unwrap returns underlying cause", func(t *testing.T) {
		err := apperror.NewErrInternal("code", "msg", cause)
		assert.Equal(t, cause, err.Unwrap())
		assert.ErrorIs(t, err, cause)
	})

	t.Run("Is matches by code", func(t *testing.T) {
		err1 := apperror.NewErrNotFound("code_a", "msg 1")
		err2 := apperror.NewErrNotFound("code_a", "different msg")
		err3 := apperror.NewErrNotFound("code_b", "msg 1")

		assert.True(t, errors.Is(err1, err2))
		assert.False(t, errors.Is(err1, err3))
		assert.False(t, errors.Is(err1, errors.New("standard error")))
	})

	t.Run("WithCause sets new cause", func(t *testing.T) {
		base := apperror.NewErrNotFound("code", "msg")
		assert.Nil(t, base.Err)

		withCause := base.WithCause(cause)
		assert.Equal(t, cause, withCause.Err)
		assert.Equal(t, base.Code, withCause.Code)
		assert.Equal(t, base.Kind, withCause.Kind)
		assert.Equal(t, base.Message, withCause.Message)
	})

	t.Run("IsDomain on AppError", func(t *testing.T) {
		var nilAppErr *apperror.AppError
		assert.False(t, nilAppErr.IsDomain())

		notFound := apperror.NewErrNotFound("not_found", "missing")
		assert.True(t, notFound.IsDomain())

		internal := apperror.NewErrInternal("internal", "boom", nil)
		assert.False(t, internal.IsDomain())
	})

	t.Run("Constructors create expected kinds", func(t *testing.T) {
		assert.Equal(
			t,
			apperror.ErrKindNotFound,
			apperror.NewErrNotFound("c", "m").Kind,
		)
		assert.Equal(
			t,
			apperror.ErrKindBadRequest,
			apperror.NewErrBadRequest("c", "m").Kind,
		)
		assert.Equal(
			t,
			apperror.ErrKindConflict,
			apperror.NewErrConflict("c", "m").Kind,
		)
		assert.Equal(
			t,
			apperror.ErrKindInternal,
			apperror.NewErrInternal("c", "m", nil).Kind,
		)
		assert.Equal(
			t,
			apperror.ErrKindUnavailable,
			apperror.NewErrUnavailable("c", "m", nil).Kind,
		)
		assert.Equal(
			t,
			apperror.ErrKindExternalService,
			apperror.NewErrExternalService("c", "m", nil).Kind,
		)
	})

	t.Run("Web error constructors", func(t *testing.T) {
		webErr := apperror.NewErrWeb("custom_web", "custom message")
		assert.Equal(t, apperror.ErrKindBadRequest, webErr.Kind)
		assert.Equal(t, "custom_web", webErr.Code)
		assert.Equal(t, "custom message", webErr.Message)

		bodyErr := apperror.NewErrRequestBodyMalformed()
		assert.Equal(t, apperror.ErrKindBadRequest, bodyErr.Kind)
		assert.Equal(t, "body_malformed", bodyErr.Code)
		assert.Equal(t, "request body malformed", bodyErr.Message)

		paramsErr := apperror.NewErrRequestParamsMalformed()
		assert.Equal(t, apperror.ErrKindBadRequest, paramsErr.Kind)
		assert.Equal(t, "params_malformed", paramsErr.Code)
		assert.Equal(t, "request params malformed", paramsErr.Message)
	})
}
