package syncoutbox

import (
	"context"
	"time"

	"github.com/bryanfks-dev/logos/infra/db/correlation"
	"github.com/bryanfks-dev/logos/pkg/entity"
	entityerror "github.com/bryanfks-dev/logos/pkg/error/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (r *syncOutboxRepository) conn(ctx context.Context) *gorm.DB {
	if tx := correlation.GetTx(ctx); tx != nil {
		return tx.WithContext(ctx)
	}

	return r.db.WithContext(ctx)
}

func (r *syncOutboxRepository) Enqueue(
	ctx context.Context,
	outbox *entity.SyncOutbox,
) error {
	r.log.Info("enqueue started", zap.Any("param", outbox))

	if err := r.conn(ctx).Create(outbox).Error; err != nil {
		return entityerror.ErrSyncOutboxEnqueueFailed(err)
	}

	r.log.Info("enqueue completed", zap.Any("param", outbox))

	return nil
}

func (r *syncOutboxRepository) GetPendingBatch(
	ctx context.Context,
	limit int,
) ([]*entity.SyncOutbox, error) {
	r.log.Info("get pending batch started", zap.Any("param", limit))

	var outboxes []*entity.SyncOutbox

	if err := r.conn(ctx).Order("queued_at ASC").Limit(limit).Find(&outboxes).Error; err != nil {
		return nil, entityerror.ErrSyncOutboxRetrievalFailed(err)
	}

	r.log.Info("get pending batch completed", zap.Any("param", limit))

	return outboxes, nil
}

func (r *syncOutboxRepository) MarkSynced(
	ctx context.Context,
	ids []string,
) error {
	r.log.Info("mark synced started", zap.Any("param", ids))

	outboxModel := &entity.SyncOutbox{}
	if err := r.conn(ctx).Where("id IN ?", ids).Delete(outboxModel).Error; err != nil {
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

	outboxModel := &entity.SyncOutbox{}
	if err := r.conn(ctx).Where("queued_at <= ?", before).Delete(outboxModel).Error; err != nil {
		return entityerror.ErrSyncOutboxClearSyncedFailed(err)
	}

	r.log.Info("clear synced completed", zap.Time("before", before))

	return nil
}
