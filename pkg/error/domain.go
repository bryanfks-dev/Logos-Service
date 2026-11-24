package apperror

func NewErrNotFound(code, message string) *AppError {
	return &AppError{
		Kind:    ErrKindNotFound,
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

func NewErrMethodNotAllowed(code, message string) *AppError {
	return &AppError{
		Kind:    ErrKindMethodNotAllowed,
		Code:    code,
		Message: message,
	}
}
