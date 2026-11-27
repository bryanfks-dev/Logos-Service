package preacher

import (
	"context"
	"time"

	"github.com/bryanfks-dev/logos/pkg/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PreacherRepository interface {
	Create(ctx context.Context, preacher *entity.Preacher) error
	CreateBatch(ctx context.Context, preachers []*entity.Preacher) error
	Update(ctx context.Context, preacher *entity.Preacher) error
	UpdateBatch(ctx context.Context, preachers []*entity.Preacher) error
	FindByID(ctx context.Context, id string) (*entity.Preacher, error)
	FindByIDs(ctx context.Context, ids []string) ([]*entity.Preacher, error)
	FindAll(
		ctx context.Context,
		filter entity.PreacherFilter,
	) ([]*entity.Preacher, error)
	FindUpdatedSince(
		ctx context.Context,
		since time.Time,
	) ([]*entity.Preacher, error)
	Delete(ctx context.Context, id string) error
}

type preacherRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewPreacherRepository(db *gorm.DB, log *zap.Logger) PreacherRepository {
	return &preacherRepository{
		db:  db,
		log: log,
	}
}
