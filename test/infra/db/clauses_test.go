package db_test

import (
	"context"
	"testing"

	"github.com/bryanfks-dev/logos/infra/db"
	"github.com/bryanfks-dev/logos/infra/db/correlation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestWithRowLock(t *testing.T) {
	testDB, err := gorm.Open(
		sqlite.Open("file::memory:?cache=shared"),
		&gorm.Config{},
	)
	require.NoError(t, err)

	t.Run("returns nil when query is nil", func(t *testing.T) {
		assert.Nil(t, db.WithRowLock(context.Background(), nil))
	})

	t.Run(
		"does not add locking clause without transaction",
		func(t *testing.T) {
			query := testDB.Model(&TestUser{})
			locked := db.WithRowLock(context.Background(), query)

			_, hasLock := locked.Statement.Clauses["FOR"]
			assert.False(t, hasLock)
		},
	)

	t.Run(
		"adds FOR UPDATE locking clause inside transaction",
		func(t *testing.T) {
			ctx := correlation.WithTx(context.Background(), testDB)
			query := testDB.Model(&TestUser{})
			locked := db.WithRowLock(ctx, query)

			_, hasLock := locked.Statement.Clauses["FOR"]
			assert.True(t, hasLock)
		},
	)
}
