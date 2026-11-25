package db_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/bryanfks-dev/logos/db"
	"github.com/bryanfks-dev/logos/db/correlation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TestUser struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"column:name"`
}

func setupTxTestDB(t *testing.T) *gorm.DB {
	testDB, err := gorm.Open(
		sqlite.Open("file::memory:?cache=shared"),
		&gorm.Config{},
	)
	require.NoError(t, err)
	err = testDB.AutoMigrate(&TestUser{})
	require.NoError(t, err)
	return testDB
}

func TestTxManager(t *testing.T) {
	testDB := setupTxTestDB(t)
	txManager := db.NewTxManager(testDB)

	t.Run("nil manager or nil db returns error", func(t *testing.T) {
		var nilMgr *db.TxManager
		err := nilMgr.Do(
			context.Background(),
			func(ctx context.Context) error { return nil },
		)
		assert.Error(t, err)
		assert.Equal(t, "db is nil", err.Error())

		nilDBMgr := db.NewTxManager(nil)
		err = nilDBMgr.Do(
			context.Background(),
			func(ctx context.Context) error { return nil },
		)
		assert.Error(t, err)
		assert.Equal(t, "db is nil", err.Error())
	})

	t.Run("nil fn returns nil", func(t *testing.T) {
		err := txManager.Do(context.Background(), nil)
		assert.NoError(t, err)
	})

	t.Run("successful transaction commits", func(t *testing.T) {
		ctx := context.Background()
		err := txManager.Do(ctx, func(txCtx context.Context) error {
			tx := correlation.Tx(txCtx)
			require.NotNil(t, tx)
			return tx.Create(&TestUser{Name: "Alice"}).Error
		})
		assert.NoError(t, err)

		var user TestUser
		err = testDB.First(&user, "name = ?", "Alice").Error
		assert.NoError(t, err)
		assert.Equal(t, "Alice", user.Name)
	})

	t.Run("error in transaction rolls back", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("business error")
		err := txManager.Do(ctx, func(txCtx context.Context) error {
			tx := correlation.Tx(txCtx)
			require.NotNil(t, tx)
			if err := tx.Create(&TestUser{Name: "Bob"}).Error; err != nil {
				return err
			}
			return expectedErr
		})
		assert.ErrorIs(t, err, expectedErr)

		var user TestUser
		err = testDB.First(&user, "name = ?", "Bob").Error
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run(
		"nested transaction reuses existing tx from context",
		func(t *testing.T) {
			ctx := context.Background()
			outerExecuted := false
			innerExecuted := false

			err := txManager.Do(ctx, func(outerCtx context.Context) error {
				outerExecuted = true
				outerTx := correlation.Tx(outerCtx)
				require.NotNil(t, outerTx)

				return txManager.Do(
					outerCtx,
					func(innerCtx context.Context) error {
						innerExecuted = true
						innerTx := correlation.Tx(innerCtx)
						assert.Equal(t, outerTx, innerTx)
						return innerTx.Create(&TestUser{Name: "Charlie"}).Error
					},
				)
			})
			assert.NoError(t, err)
			assert.True(t, outerExecuted)
			assert.True(t, innerExecuted)

			var user TestUser
			err = testDB.First(&user, "name = ?", "Charlie").Error
			assert.NoError(t, err)
		},
	)

	t.Run("DoWithOptions passes options", func(t *testing.T) {
		ctx := context.Background()
		opts := &sql.TxOptions{
			ReadOnly: false,
		}
		err := txManager.DoWithOptions(
			ctx,
			opts,
			func(txCtx context.Context) error {
				tx := correlation.Tx(txCtx)
				require.NotNil(t, tx)
				return nil
			},
		)
		assert.NoError(t, err)
	})
}
