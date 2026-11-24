package apperror

import (
	"errors"
	"slices"
)

type ErrKind int

const (
	ErrKindNotFound ErrKind = iota
	ErrKindConflict
	ErrKindMethodNotAllowed
	ErrKindUnavailable
	ErrKindExternalService
	ErrKindInternal
)

const ErrKindUnvailable = ErrKindUnavailable

var domainErrKind = []ErrKind{
	ErrKindNotFound,
	ErrKindConflict,
	ErrKindMethodNotAllowed,
}

func (k ErrKind) IsDomain() bool {
	return slices.Contains(domainErrKind, k)
}

type AppError struct {
	Kind    ErrKind
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Code + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	return ok && t.Code == e.Code
}

func (e *AppError) WithCause(cause error) *AppError {
	return &AppError{
		Kind:    e.Kind,
		Code:    e.Code,
		Message: e.Message,
		Err:     cause,
	}
}

func KindOf(err error) (ErrKind, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Kind, true
	}
	return 0, false
}
