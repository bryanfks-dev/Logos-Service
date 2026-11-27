package tag

import (
	"context"
	"time"

	"github.com/bryanfks-dev/logos/pkg/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TagRepository interface {
	Create(ctx context.Context, tag *entity.Tag) error
	CreateBatch(ctx context.Context, tags []*entity.Tag) error
	Update(ctx context.Context, tag *entity.Tag) error
	UpdateBatch(ctx context.Context, tags []*entity.Tag) error
	FindByID(ctx context.Context, id string) (*entity.Tag, error)
	FindByIDsOrNames(
		ctx context.Context,
		ids []string,
		names []string,
	) ([]*entity.Tag, error)
	FindAll(ctx context.Context) ([]*entity.Tag, error)
	FindUpdatedSince(
		ctx context.Context,
		since time.Time,
	) ([]*entity.Tag, error)
	Delete(ctx context.Context, id string) error
}

type tagRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewTagRepository(db *gorm.DB, log *zap.Logger) TagRepository {
	return &tagRepository{
		db:  db,
		log: log,
	}
}
