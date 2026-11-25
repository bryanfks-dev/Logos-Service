package converter_test

import (
	"testing"
	"time"

	"github.com/bryanfks-dev/logos/internal/model/converter"
	"github.com/bryanfks-dev/logos/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSermonSeriesToResponse(t *testing.T) {
	now := time.Now()

	t.Run("nil series returns nil", func(t *testing.T) {
		assert.Nil(t, converter.SermonSeriesToResponse(nil))
	})

	t.Run("converts series", func(t *testing.T) {
		s := &entity.SermonSeries{
			ID:          "s-1",
			Name:        "Genesis",
			Description: "Origins",
			CreatedAt:   now,
			UpdatedAt:   now,
			DeletedAt:   gorm.DeletedAt{Time: now, Valid: true},
		}
		res := converter.SermonSeriesToResponse(s)
		require.NotNil(t, res)
		assert.Equal(t, "s-1", res.ID)
		assert.Equal(t, "Genesis", res.Name)
		assert.NotNil(t, res.DeletedAt)
	})
}
