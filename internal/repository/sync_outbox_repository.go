package repository

import (
	"context"
	"time"

	"github.com/bryanfks-dev/selah/db/correlation"
	"github.com/bryanfks-dev/selah/pkg/entity"
	entityerror "github.com/bryanfks-dev/selah/pkg/error/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SyncOutboxRepository interface {
	Enqueue(ctx context.Context, entry *entity.SyncOutboxEntry) error
	GetPendingBatch(
		ctx context.Context,
		limit int,
	) ([]entity.SyncOutboxEntry, error)
	MarkSynced(ctx context.Context, ids []string) error
	ClearSynced(ctx context.Context, before time.Time) error
}

type syncOutboxRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewSyncOutboxRepository(
	db *gorm.DB,
	log *zap.Logger,
) SyncOutboxRepository {
	return &syncOutboxRepository{
		db:  db,
		log: log,
	}
}

func (r *syncOutboxRepository) conn(ctx context.Context) *gorm.DB {
	if tx := correlation.Tx(ctx); tx != nil {
		return tx.WithContext(ctx)
	}

	return r.db.WithContext(ctx)
}

func (r *syncOutboxRepository) Enqueue(
	ctx context.Context,
	entry *entity.SyncOutboxEntry,
) error {
	r.log.Info("enqueue started", zap.Any("param", entry))

	if err := r.conn(ctx).Create(entry).Error; err != nil {
		r.log.Error("enqueue failed", zap.Error(err), zap.Any("param", entry))

		return entityerror.ErrSyncOutboxEnqueueFailed(err)
	}

	r.log.Info("enqueue completed", zap.Any("param", entry))

	return nil
}

func (r *syncOutboxRepository) GetPendingBatch(
	ctx context.Context,
	limit int,
) ([]entity.SyncOutboxEntry, error) {
	r.log.Info("get pending batch started", zap.Any("param", limit))

	var entries []entity.SyncOutboxEntry

	if err := r.conn(ctx).Order("queued_at ASC").Limit(limit).Find(&entries).Error; err != nil {
		r.log.Error(
			"get pending batch failed",
			zap.Error(err),
			zap.Any("param", limit),
		)

		return nil, entityerror.ErrSyncOutboxRetrievalFailed(err)
	}

	r.log.Info("get pending batch completed", zap.Any("param", limit))

	return entries, nil
}

func (r *syncOutboxRepository) MarkSynced(
	ctx context.Context,
	ids []string,
) error {
	r.log.Info("mark synced started", zap.Any("param", ids))

	if err := r.conn(ctx).Where("id IN ?", ids).Delete(&entity.SyncOutboxEntry{}).Error; err != nil {
		r.log.Error("mark synced failed", zap.Error(err), zap.Any("param", ids))

		return entityerror.ErrSyncOutboxMarkSyncedFailed(err)
	}

	r.log.Info("mark synced completed", zap.Any("param", ids))

	return nil
}

func (r *syncOutboxRepository) ClearSynced(
	ctx context.Context,
	before time.Time,
) error {
	r.log.Info("clear synced started", zap.Time("before", before))

	if err := r.conn(ctx).Where("queued_at <= ?", before).Delete(&entity.SyncOutboxEntry{}).Error; err != nil {
		r.log.Error(
			"clear synced failed",
			zap.Error(err),
			zap.Time("before", before),
		)

		return entityerror.ErrSyncOutboxClearSyncedFailed(err)
	}

	r.log.Info("clear synced completed", zap.Time("before", before))

	return nil
}
