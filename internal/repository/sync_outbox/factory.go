package syncoutbox

import (
	"context"
	"time"

	"github.com/bryanfks-dev/logos/pkg/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SyncOutboxRepository interface {
	Enqueue(ctx context.Context, outbox *entity.SyncOutbox) error
	GetPendingBatch(
		ctx context.Context,
		limit int,
	) ([]*entity.SyncOutbox, error)
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
