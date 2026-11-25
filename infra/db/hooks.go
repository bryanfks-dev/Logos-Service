package db

import (
	"context"
	"sync"
)

const afterCommitKey string = "after_commit"

type txHookList struct {
	mu    sync.Mutex
	hooks []func() error
}

func WithAfterCommit(ctx context.Context) (context.Context, func() error) {
	list := &txHookList{}
	hookCtx := context.WithValue(ctx, afterCommitKey, list)

	hookFn := func() error {
		list.mu.Lock()
		hooks := make([]func() error, len(list.hooks))
		copy(hooks, list.hooks)
		list.mu.Unlock()

		for _, hook := range hooks {
			if err := hook(); err != nil {
				return err
			}
		}

		return nil
	}

	return hookCtx, hookFn
}

func AddAfterCommit(ctx context.Context, hookFn func() error) bool {
	if hookFn == nil {
		return false
	}

	list, ok := ctx.Value(afterCommitKey).(*txHookList)
	if !ok || list == nil {
		return false
	}

	list.mu.Lock()
	list.hooks = append(list.hooks, hookFn)
	list.mu.Unlock()

	return true
}
