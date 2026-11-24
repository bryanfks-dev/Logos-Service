package dbcorrelation_test

import (
	"context"
	"testing"

	dbcorrelation "github.com/bryanfks-dev/selah/infrastructure/db/correlation"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestWithTx_And_GetTx(t *testing.T) {
	db, err := gorm.Open(
		sqlite.Open("file::memory:?cache=shared"),
		&gorm.Config{},
	)
	assert.NoError(t, err)

	t.Run(
		"with nil context should create background context",
		func(t *testing.T) {
			ctx := dbcorrelation.WithTx(nil, db)
			assert.NotNil(t, ctx)
			assert.Equal(t, db, dbcorrelation.GetTx(ctx))
		},
	)

	t.Run("with standard context", func(t *testing.T) {
		ctx := context.Background()
		assert.Nil(t, dbcorrelation.GetTx(ctx))

		ctxWithTx := dbcorrelation.WithTx(ctx, db)
		assert.Equal(t, db, dbcorrelation.GetTx(ctxWithTx))
	})

	t.Run("get tx from nil context", func(t *testing.T) {
		assert.Nil(t, dbcorrelation.GetTx(nil))
	})

	t.Run("get tx with wrong type in context", func(t *testing.T) {
		type customKey string
		ctx := context.WithValue(
			context.Background(),
			customKey("tx"),
			"not-a-gorm-db",
		)
		assert.Nil(t, dbcorrelation.GetTx(ctx))
	})
}
