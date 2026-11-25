package builder

import (
	"strings"
	"time"

	"github.com/bryanfks-dev/logos/internal/model"
)

type APIResponseBuilder[T any] struct {
	success bool
	data    *T
	reqID   string
	err     *model.APIError
}

func NewAPIResponseBuilder[T any]() *APIResponseBuilder[T] {
	return &APIResponseBuilder[T]{}
}

func (b *APIResponseBuilder[T]) SetError(
	err model.APIError,
) *APIResponseBuilder[T] {
	b.success = false
	b.err = &err

	return b
}

func (b *APIResponseBuilder[T]) SetData(data *T) *APIResponseBuilder[T] {
	b.data = data

	return b
}

func (b *APIResponseBuilder[T]) SetRequestID(
	reqID string,
) *APIResponseBuilder[T] {
	b.reqID = reqID

	return b
}

func (b *APIResponseBuilder[T]) Build() *model.APIResponse[T] {
	if strings.TrimSpace(b.reqID) == "" {
		panic("`requestID` must not blank")
	}

	success := b.err == nil
	meta := model.APIMeta{
		RequestID: b.reqID,
		Timestamp: time.Now().UTC(),
	}

	return &model.APIResponse[T]{
		Success: success,
		Data:    b.data,
		Error:   b.err,
		Meta:    meta,
	}
}
