package converter_test

import (
	"testing"
	"time"

	"github.com/bryanfks-dev/logos/internal/model/converter"
	"github.com/bryanfks-dev/logos/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagToResponse(t *testing.T) {
	now := time.Now().UTC()

	t.Run("nil tag returns nil", func(t *testing.T) {
		assert.Nil(t, converter.TagToResponse(nil))
	})

	t.Run("converts tag", func(t *testing.T) {
		tag := &entity.Tag{
			ID:         "t-1",
			Name:       "Grace",
			UsageCount: 5,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		res := converter.TagToResponse(tag)
		require.NotNil(t, res)
		assert.Equal(t, "t-1", res.ID)
		assert.Equal(t, "Grace", res.Name)
		assert.Equal(t, 5, res.UsageCount)
	})
}
