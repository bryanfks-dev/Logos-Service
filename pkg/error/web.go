package apperror

func NewErrWeb(code, message string) *AppError {
	return &AppError{
		Kind:    ErrKindBadRequest,
		Code:    code,
		Message: message,
	}
}

func NewErrRequestBodyMalformed() *AppError {
	return NewErrWeb("body_malformed", "request body malformed")
}

func NewErrRequestParamsMalformed() *AppError {
	return NewErrWeb("params_malformed", "request params malformed")
}
