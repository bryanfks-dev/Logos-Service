package gathering

import (
	"context"
	"time"

	"github.com/bryanfks-dev/logos/pkg/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GatheringRepository interface {
	Create(ctx context.Context, gathering *entity.Gathering) error
	CreateBatch(ctx context.Context, gatherings []*entity.Gathering) error
	Update(ctx context.Context, gathering *entity.Gathering) error
	UpdateBatch(ctx context.Context, gatherings []*entity.Gathering) error
	FindByID(ctx context.Context, id string) (*entity.Gathering, error)
	FindByIDs(ctx context.Context, ids []string) ([]*entity.Gathering, error)
	FindAll(ctx context.Context) ([]*entity.Gathering, error)
	FindUpdatedSince(
		ctx context.Context,
		since time.Time,
	) ([]*entity.Gathering, error)
	Delete(ctx context.Context, id string) error
}

type gatheringRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewGatheringRepository(db *gorm.DB, log *zap.Logger) GatheringRepository {
	return &gatheringRepository{
		db:  db,
		log: log,
	}
}
