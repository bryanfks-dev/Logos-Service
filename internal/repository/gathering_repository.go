package repository

import (
	"context"
	"errors"

	"github.com/bryanfks-dev/selah/db/correlation"
	"github.com/bryanfks-dev/selah/pkg/entity"
	entityerror "github.com/bryanfks-dev/selah/pkg/error/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GatheringRepository interface {
	Create(ctx context.Context, gathering *entity.Gathering) error
	Update(ctx context.Context, gathering *entity.Gathering) error
	FindByID(ctx context.Context, id string) (*entity.Gathering, error)
	FindAll(ctx context.Context) ([]entity.Gathering, error)
	FindUpdatedSince(
		ctx context.Context,
		since int64,
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

func (r *gatheringRepository) conn(ctx context.Context) *gorm.DB {
	if tx := correlation.Tx(ctx); tx != nil {
		return tx.WithContext(ctx)
	}

	return r.db.WithContext(ctx)
}

func (r *gatheringRepository) Create(
	ctx context.Context,
	gathering *entity.Gathering,
) error {
	r.log.Info("create started", zap.Any("param", gathering))

	if err := r.conn(ctx).Create(gathering).Error; err != nil {
		r.log.Error(
			"create failed",
			zap.Error(err),
			zap.Any("param", gathering),
		)

		return entityerror.ErrGatheringCreationFailed(err)
	}

	r.log.Info("create completed", zap.Any("param", gathering))

	return nil
}

func (r *gatheringRepository) Update(
	ctx context.Context,
	gathering *entity.Gathering,
) error {
	r.log.Info("update started", zap.Any("param", gathering))

	if err := r.conn(ctx).Save(gathering).Error; err != nil {
		r.log.Error(
			"update failed",
			zap.Error(err),
			zap.Any("param", gathering),
		)

		return entityerror.ErrGatheringUpdateFailed(err)
	}

	r.log.Info("update completed", zap.Any("param", gathering))

	return nil
}

func (r *gatheringRepository) FindByID(
	ctx context.Context,
	id string,
) (*entity.Gathering, error) {
	r.log.Info("find by id started", zap.Any("param", id))

	var item entity.Gathering

	if err := r.conn(ctx).Where("id = ?", id).First(&item).Error; err != nil {
		r.log.Error("find by id failed", zap.Error(err), zap.Any("param", id))

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entityerror.ErrGatheringNotFound(id)
		}

		return nil, entityerror.ErrGatheringRetrievalFailed(err)
	}

	r.log.Info("find by id completed", zap.Any("param", id))

	return &item, nil
}

func (r *gatheringRepository) FindAll(
	ctx context.Context,
) ([]entity.Gathering, error) {
	r.log.Info("find all started", zap.Any("param", nil))

	var items []entity.Gathering

	if err := r.conn(ctx).Order("sort_order ASC").Find(&items).Error; err != nil {
		r.log.Error("find all failed", zap.Error(err), zap.Any("param", nil))

		return nil, entityerror.ErrGatheringRetrievalFailed(err)
	}

	r.log.Info("find all completed", zap.Any("param", nil))

	return items, nil
}

func (r *gatheringRepository) FindUpdatedSince(
	ctx context.Context,
	since int64,
) ([]*entity.Gathering, error) {
	r.log.Info("find updated since started", zap.Int64("since", since))

	sinceTime := parseSinceTime(since)
	var items []*entity.Gathering

	err := r.conn(ctx).
		Unscoped().
		Where("updated_at > ?", sinceTime).
		Order("updated_at ASC").
		Find(&items).Error
	if err != nil {
		r.log.Error(
			"find updated since failed",
			zap.Error(err),
			zap.Int64("since", since),
		)

		return nil, entityerror.ErrGatheringRetrievalFailed(err)
	}

	if items == nil {
		items = make([]*entity.Gathering, 0)
	}

	r.log.Info("find updated since completed", zap.Int("count", len(items)))

	return items, nil
}

func (r *gatheringRepository) Delete(ctx context.Context, id string) error {
	r.log.Info("delete started", zap.Any("param", id))

	if err := r.conn(ctx).Where("id = ?", id).Delete(&entity.Gathering{}).Error; err != nil {
		r.log.Error("delete failed", zap.Error(err), zap.Any("param", id))

		return entityerror.ErrGatheringDeleteFailed(err)
	}

	r.log.Info("delete completed", zap.Any("param", id))

	return nil
}
