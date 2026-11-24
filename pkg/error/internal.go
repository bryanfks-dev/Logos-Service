package apperror

func NewErrInternal(code, message string, cause error) *AppError {
	return &AppError{
		Kind:    ErrKindInternal,
		Code:    code,
		Message: message,
		Err:     cause,
	}
}

func NewErrUnavailable(code, message string, cause error) *AppError {
	return &AppError{
		Kind:    ErrKindUnavailable,
		Code:    code,
		Message: message,
		Err:     cause,
	}
}

func NewErrExternalService(code, message string, cause error) *AppError {
	return &AppError{
		Kind:    ErrKindExternalService,
		Code:    code,
		Message: message,
		Err:     cause,
	}
}
