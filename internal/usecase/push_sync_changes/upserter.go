package pushsyncchanges

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type lastWriteWinsUpserter[T any] struct {
	id           func(*T) string
	updatedAt    func(*T) time.Time
	beforeUpdate func(incoming *T, existing *T)
	findByIDs    func(context.Context, []string) ([]*T, error)
	createBatch  func(context.Context, []*T) error
	updateBatch  func(context.Context, []*T) error
	log          *zap.Logger
}

func (w *lastWriteWinsUpserter[T]) upsert(
	ctx context.Context,
	items []*T,
) error {
	newItems, ids, incomingByID := w.coalesce(items)

	toCreate := make([]*T, 0, len(newItems))
	toCreate = append(toCreate, newItems...)
	toUpdate := make([]*T, 0, len(ids))

	if len(ids) > 0 {
		existingList, err := w.findByIDs(ctx, ids)
		if err != nil {
			return err
		}

		existingByID := make(map[string]*T, len(existingList))
		for _, existing := range existingList {
			if existing == nil {
				continue
			}

			existingByID[w.id(existing)] = existing
		}

		for _, id := range ids {
			item := incomingByID[id]
			existing := existingByID[id]

			if existing == nil {
				toCreate = append(toCreate, item)
				continue
			}

			if w.updatedAt(item).Before(w.updatedAt(existing)) {
				w.log.Debug(
					"skipped stale update",
					zap.String("id", w.id(item)),
				)

				continue
			}

			if w.beforeUpdate != nil {
				w.beforeUpdate(item, existing)
			}

			toUpdate = append(toUpdate, item)
		}
	}

	if err := w.createBatch(ctx, toCreate); err != nil {
		return err
	}

	return w.updateBatch(ctx, toUpdate)
}

func (w *lastWriteWinsUpserter[T]) coalesce(
	items []*T,
) ([]*T, []string, map[string]*T) {
	newItems := make([]*T, 0)
	ids := make([]string, 0, len(items))
	byID := make(map[string]*T, len(items))

	for _, item := range items {
		if item == nil {
			continue
		}

		id := w.id(item)
		if id == "" {
			newItems = append(newItems, item)
			continue
		}

		prev, exists := byID[id]
		if !exists {
			ids = append(ids, id)
			byID[id] = item
			continue
		}

		if w.updatedAt(item).Before(w.updatedAt(prev)) {
			w.log.Debug(
				"skipped stale duplicate item in batch",
				zap.String("id", id),
			)

			continue
		}

		byID[id] = item
	}

	return newItems, ids, byID
}
