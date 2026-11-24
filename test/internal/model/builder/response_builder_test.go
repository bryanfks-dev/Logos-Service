package modelbuilder_test

import (
	"testing"

	"github.com/bryanfks-dev/selah/internal/model"
	modelbuilder "github.com/bryanfks-dev/selah/internal/model/builder"
	"github.com/stretchr/testify/assert"
)

func TestApiResponseBuilder(t *testing.T) {
	t.Run("panics when requestID is empty or blank", func(t *testing.T) {
		assert.PanicsWithValue(t, "`requestID` must not blank", func() {
			modelbuilder.NewApiResponseBuilder[string]().Build()
		})

		assert.PanicsWithValue(t, "`requestID` must not blank", func() {
			modelbuilder.NewApiResponseBuilder[string]().
				SetRequestId("   ").
				Build()
		})
	})

	t.Run("builds successful response", func(t *testing.T) {
		res := modelbuilder.NewApiResponseBuilder[string]().
			SetRequestId("req-success-123").
			Build()

		assert.True(t, res.Success)
		assert.Nil(t, res.Error)
		assert.Nil(t, res.Data)
		assert.Equal(t, "req-success-123", res.Meta.RequestID)
		assert.False(t, res.Meta.Timestamp.IsZero())
	})

	t.Run("builds successful response with data", func(t *testing.T) {
		res := modelbuilder.NewApiResponseBuilder[string]().
			SetRequestId("req-success-123").
			SetData("hello world").
			Build()

		assert.True(t, res.Success)
		assert.Nil(t, res.Error)
		assert.NotNil(t, res.Data)
		assert.Equal(t, "hello world", *res.Data)
		assert.Equal(t, "req-success-123", res.Meta.RequestID)
		assert.False(t, res.Meta.Timestamp.IsZero())
	})

	t.Run("builds error response", func(t *testing.T) {
		apiErr := model.ApiError{
			Code: "test_error",
			Fields: []model.FieldError{
				{Field: "email", Message: "invalid email format"},
			},
		}

		res := modelbuilder.NewApiResponseBuilder[any]().
			SetRequestId("req-err-456").
			SetError(apiErr).
			Build()

		assert.False(t, res.Success)
		assert.NotNil(t, res.Error)
		assert.Equal(t, "test_error", res.Error.Code)
		assert.Len(t, res.Error.Fields, 1)
		assert.Equal(t, "email", res.Error.Fields[0].Field)
		assert.Equal(t, "req-err-456", res.Meta.RequestID)
	})

	t.Run(
		"builds response using SetRequestID with proper acronym casing",
		func(t *testing.T) {
			res := modelbuilder.NewApiResponseBuilder[string]().
				SetRequestID("req-case-789").
				Build()

			assert.Equal(t, "req-case-789", res.Meta.RequestID)
		},
	)
}
