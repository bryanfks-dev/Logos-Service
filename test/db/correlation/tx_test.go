package correlation_test

import (
	"context"
	"testing"

	"github.com/bryanfks-dev/selah/db/correlation"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestWithTx_And_Tx(t *testing.T) {
	db, err := gorm.Open(
		sqlite.Open("file::memory:?cache=shared"),
		&gorm.Config{},
	)
	assert.NoError(t, err)

	t.Run("with standard context", func(t *testing.T) {
		ctx := context.Background()
		assert.Nil(t, correlation.Tx(ctx))

		ctxWithTx := correlation.WithTx(ctx, db)
		assert.Equal(t, db, correlation.Tx(ctxWithTx))
		assert.Equal(t, db, correlation.GetTx(ctxWithTx))
	})

	t.Run("get tx with wrong type in context", func(t *testing.T) {
		type customKey string
		ctx := context.WithValue(
			context.Background(),
			customKey("tx"),
			"not-a-gorm-db",
		)
		assert.Nil(t, correlation.Tx(ctx))
	})
}
