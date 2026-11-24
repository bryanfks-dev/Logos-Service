package builder_test

import (
	"testing"

	"github.com/bryanfks-dev/selah/internal/model"
	"github.com/bryanfks-dev/selah/internal/model/builder"
	"github.com/stretchr/testify/assert"
)

func TestAPIResponseBuilder(t *testing.T) {
	t.Run("panics when requestID is empty or blank", func(t *testing.T) {
		assert.PanicsWithValue(t, "`requestID` must not blank", func() {
			builder.NewAPIResponseBuilder[string]().Build()
		})

		assert.PanicsWithValue(t, "`requestID` must not blank", func() {
			builder.NewAPIResponseBuilder[string]().
				SetRequestID("   ").
				Build()
		})
	})

	t.Run(
		"builds successful response with data and requestID",
		func(t *testing.T) {
			res := builder.NewAPIResponseBuilder[string]().
				SetRequestID("req-success-123").
				SetData("hello world").
				Build()

			assert.NotNil(t, res)
			assert.True(t, res.Success)
			assert.Equal(t, "hello world", *res.Data)
			assert.Equal(t, "req-success-123", res.Meta.RequestID)
			assert.False(t, res.Meta.Timestamp.IsZero())
			assert.Nil(t, res.Error)
		},
	)

	t.Run("builds successful response without data", func(t *testing.T) {
		res := builder.NewAPIResponseBuilder[string]().
			SetRequestID("req-success-123").
			Build()

		assert.NotNil(t, res)
		assert.True(t, res.Success)
		assert.Nil(t, res.Data)
		assert.Equal(t, "req-success-123", res.Meta.RequestID)
		assert.False(t, res.Meta.Timestamp.IsZero())
		assert.Nil(t, res.Error)
	})

	t.Run("builds error response", func(t *testing.T) {
		apiErr := model.APIError{
			Code: "RESOURCE_NOT_FOUND",
			Errors: []model.FieldError{
				{Field: "id", Message: "id does not exist"},
			},
		}

		res := builder.NewAPIResponseBuilder[any]().
			SetRequestID("req-err-456").
			SetError(apiErr).
			Build()

		assert.NotNil(t, res)
		assert.False(t, res.Success)
		assert.Nil(t, res.Data)
		assert.Equal(t, "RESOURCE_NOT_FOUND", res.Error.Code)
		assert.Len(t, res.Error.Errors, 1)
		assert.Equal(t, "req-err-456", res.Meta.RequestID)
		assert.False(t, res.Meta.Timestamp.IsZero())
	})
}
