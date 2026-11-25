package converter_test

import (
	"testing"
	"time"

	"github.com/bryanfks-dev/logos/internal/model/converter"
	"github.com/bryanfks-dev/logos/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatheringToResponse(t *testing.T) {
	now := time.Now().UTC()

	t.Run("nil gathering returns nil", func(t *testing.T) {
		assert.Nil(t, converter.GatheringToResponse(nil))
	})

	t.Run("converts gathering", func(t *testing.T) {
		g := &entity.Gathering{
			ID:        "g-1",
			Name:      "Sunday",
			ColorHex:  "#111111",
			SortOrder: 2,
			CreatedAt: now,
			UpdatedAt: now,
		}
		res := converter.GatheringToResponse(g)
		require.NotNil(t, res)
		assert.Equal(t, "g-1", res.ID)
		assert.Equal(t, "Sunday", res.Name)
		assert.Equal(t, "#111111", res.ColorHex)
		assert.Equal(t, 2, res.SortOrder)
		assert.Nil(t, res.DeletedAt)
	})
}
