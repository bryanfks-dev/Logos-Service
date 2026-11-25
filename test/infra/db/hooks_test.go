package db_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/bryanfks-dev/logos/infra/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAfterCommitHooks(t *testing.T) {
	t.Run("returns false when no hook list in context", func(t *testing.T) {
		ctx := context.Background()
		ok := db.AddAfterCommit(ctx, func() error { return nil })
		assert.False(t, ok)
	})

	t.Run("returns false when hookFn is nil", func(t *testing.T) {
		hookCtx, _ := db.WithAfterCommit(context.Background())
		assert.False(t, db.AddAfterCommit(hookCtx, nil))
	})

	t.Run("returns false when context key has wrong type", func(t *testing.T) {
		type wrongKey string
		ctx := context.WithValue(
			context.Background(),
			wrongKey("after_commit"),
			"not-a-hook-list",
		)
		assert.False(t, db.AddAfterCommit(ctx, func() error { return nil }))
	})

	t.Run("executes registered hooks in FIFO order", func(t *testing.T) {
		hookCtx, runHooks := db.WithAfterCommit(context.Background())

		var order []int
		require.True(t, db.AddAfterCommit(hookCtx, func() error {
			order = append(order, 1)
			return nil
		}))
		require.True(t, db.AddAfterCommit(hookCtx, func() error {
			order = append(order, 2)
			return nil
		}))

		assert.NoError(t, runHooks())
		assert.Equal(t, []int{1, 2}, order)
	})

	t.Run("stops and returns error when a hook fails", func(t *testing.T) {
		hookCtx, runHooks := db.WithAfterCommit(context.Background())
		expectedErr := errors.New("hook failed")
		secondRan := false

		require.True(t, db.AddAfterCommit(hookCtx, func() error {
			return expectedErr
		}))
		require.True(t, db.AddAfterCommit(hookCtx, func() error {
			secondRan = true
			return nil
		}))

		err := runHooks()
		assert.ErrorIs(t, err, expectedErr)
		assert.False(t, secondRan)
	})

	t.Run("handles concurrent hook registration safely", func(t *testing.T) {
		hookCtx, runHooks := db.WithAfterCommit(context.Background())

		const workers = 10
		var (
			wg    sync.WaitGroup
			mu    sync.Mutex
			count int
		)

		wg.Add(workers)
		for range workers {
			go func() {
				defer wg.Done()
				ok := db.AddAfterCommit(hookCtx, func() error {
					mu.Lock()
					count++
					mu.Unlock()
					return nil
				})
				assert.True(t, ok)
			}()
		}
		wg.Wait()

		require.NoError(t, runHooks())
		assert.Equal(t, workers, count)
	})
}
