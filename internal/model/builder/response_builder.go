package builder

import (
	"strings"
	"time"

	"github.com/bryanfks-dev/selah/internal/model"
)

type ApiResponseBuilder[T any] struct {
	success   bool
	data      *T
	requestID string
	error     *model.ApiError
}

func NewApiResponseBuilder[T any]() *ApiResponseBuilder[T] {
	return &ApiResponseBuilder[T]{}
}

func (b *ApiResponseBuilder[T]) SetError(
	error model.ApiError,
) *ApiResponseBuilder[T] {
	b.success = false
	b.error = &error
	return b
}

func (b *ApiResponseBuilder[T]) SetRequestId(
	requestID string,
) *ApiResponseBuilder[T] {
	b.requestID = requestID
	return b
}

func (b *ApiResponseBuilder[T]) Build() *model.ApiResponse[T] {
	if strings.TrimSpace(b.requestID) == "" {
		panic("`requestID` must not blank")
	}

	success := b.error == nil
	meta := model.ApiMeta{
		RequestID: b.requestID,
		Timestamp: time.Now(),
	}

	return &model.ApiResponse[T]{
		Success: success,
		Data:    b.data,
		Error:   b.error,
		Meta:    meta,
	}
}
