package sermonseries

import (
	"context"
	"time"

	"github.com/bryanfks-dev/logos/pkg/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SermonSeriesRepository interface {
	Create(ctx context.Context, series *entity.SermonSeries) error
	CreateBatch(ctx context.Context, series []*entity.SermonSeries) error
	Update(ctx context.Context, series *entity.SermonSeries) error
	UpdateBatch(ctx context.Context, series []*entity.SermonSeries) error
	FindByID(ctx context.Context, id string) (*entity.SermonSeries, error)
	FindByIDs(
		ctx context.Context,
		ids []string,
	) ([]*entity.SermonSeries, error)
	FindAll(ctx context.Context) ([]*entity.SermonSeries, error)
	FindUpdatedSince(
		ctx context.Context,
		since time.Time,
	) ([]*entity.SermonSeries, error)
	Delete(ctx context.Context, id string) error
}

type sermonSeriesRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewSermonSeriesRepository(
	db *gorm.DB,
	log *zap.Logger,
) SermonSeriesRepository {
	return &sermonSeriesRepository{
		db:  db,
		log: log,
	}
}
