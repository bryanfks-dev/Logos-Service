package apperror_test

import (
	"errors"
	"testing"

	apperror "github.com/bryanfks-dev/selah/pkg/error"
	"github.com/stretchr/testify/assert"
)

func TestErrKind_IsDomain(t *testing.T) {
	assert.True(t, apperror.ErrKindNotFound.IsDomain())
	assert.True(t, apperror.ErrKindConflict.IsDomain())
	assert.True(t, apperror.ErrKindMethodNotAllowed.IsDomain())
	assert.False(t, apperror.ErrKindUnvailable.IsDomain())
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

	t.Run("KindOf", func(t *testing.T) {
		err := apperror.NewErrConflict("conflict_code", "already exists")
		kind, ok := apperror.KindOf(err)
		assert.True(t, ok)
		assert.Equal(t, apperror.ErrKindConflict, kind)

		standardErr := errors.New("regular error")
		kind, ok = apperror.KindOf(standardErr)
		assert.False(t, ok)
		assert.Equal(t, apperror.ErrKind(0), kind)
	})

	t.Run("Constructors create expected kinds", func(t *testing.T) {
		assert.Equal(
			t,
			apperror.ErrKindNotFound,
			apperror.NewErrNotFound("c", "m").Kind,
		)
		assert.Equal(
			t,
			apperror.ErrKindConflict,
			apperror.NewErrConflict("c", "m").Kind,
		)
		assert.Equal(
			t,
			apperror.ErrKindMethodNotAllowed,
			apperror.NewErrMethodNotAllowed("c", "m").Kind,
		)
		assert.Equal(
			t,
			apperror.ErrKindInternal,
			apperror.NewErrInternal("c", "m", nil).Kind,
		)
		assert.Equal(
			t,
			apperror.ErrKindUnvailable,
			apperror.NewErrUnavailable("c", "m", nil).Kind,
		)
		assert.Equal(
			t,
			apperror.ErrKindExternalService,
			apperror.NewErrExternalService("c", "m", nil).Kind,
		)
	})
}
