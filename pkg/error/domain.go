package apperror

func NewErrNotFound(code, message string) *AppError {
	return &AppError{
		Kind:    ErrKindNotFound,
		Code:    code,
		Message: message,
	}
}

func NewErrBadRequest(code, message string) *AppError {
	return &AppError{
		Kind:    ErrKindBadRequest,
		Code:    code,
		Message: message,
	}
}

func NewErrConflict(code, message string) *AppError {
	return &AppError{
		Kind:    ErrKindConflict,
		Code:    code,
		Message: message,
	}
}
