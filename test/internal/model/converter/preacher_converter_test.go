package converter_test

import (
	"testing"
	"time"

	"github.com/bryanfks-dev/logos/internal/model/converter"
	"github.com/bryanfks-dev/logos/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreacherToResponse(t *testing.T) {
	now := time.Now()

	t.Run("nil preacher returns nil", func(t *testing.T) {
		assert.Nil(t, converter.PreacherToResponse(nil))
	})

	t.Run("converts preacher", func(t *testing.T) {
		p := &entity.Preacher{
			ID:         "p-1",
			Name:       "John",
			ChurchName: "Church",
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		res := converter.PreacherToResponse(p)
		require.NotNil(t, res)
		assert.Equal(t, "p-1", res.ID)
		assert.Equal(t, "John", res.Name)
		assert.Nil(t, res.DeletedAt)
	})
}
