package builder

import (
	"strings"
	"time"

	"github.com/bryanfks-dev/selah/internal/model"
)

type APIResponseBuilder[T any] struct {
	success   bool
	data      *T
	requestID string
	error     *model.APIError
}

func NewAPIResponseBuilder[T any]() *APIResponseBuilder[T] {
	return &APIResponseBuilder[T]{}
}

func (b *APIResponseBuilder[T]) SetError(
	err model.APIError,
) *APIResponseBuilder[T] {
	b.success = false
	b.error = &err

	return b
}

func (b *APIResponseBuilder[T]) SetData(data T) *APIResponseBuilder[T] {
	b.data = &data

	return b
}

func (b *APIResponseBuilder[T]) SetRequestID(
	requestID string,
) *APIResponseBuilder[T] {
	b.requestID = requestID

	return b
}

func (b *APIResponseBuilder[T]) Build() *model.APIResponse[T] {
	if strings.TrimSpace(b.requestID) == "" {
		panic("`requestID` must not blank")
	}

	success := b.error == nil
	meta := model.APIMeta{
		RequestID: b.requestID,
		Timestamp: time.Now(),
	}

	return &model.APIResponse[T]{
		Success: success,
		Data:    b.data,
		Error:   b.error,
		Meta:    meta,
	}
}
